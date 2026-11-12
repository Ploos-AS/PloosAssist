//go:build linux && pipewire && cgo

package pipewirecapture

/*
#cgo pkg-config: libpipewire-0.3
#include <pipewire/pipewire.h>
#include <spa/param/video/format-utils.h>
#include <spa/param/video/raw.h>
#include <spa/pod/builder.h>
#include <spa/buffer/buffer.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <errno.h>

#define PA_LIMIT (16u * 1024u * 1024u)

struct pa_capture {
 struct pw_thread_loop *loop;
 struct pw_context *context;
 struct pw_core *core;
 struct pw_stream *stream;
 struct spa_hook listener;
 unsigned char *pixels;
 size_t size;
 uint32_t width, height, stride;
 int received;
};

static void pa_process(void *userdata) {
 struct pa_capture *c = userdata;
 struct pw_buffer *b = pw_stream_dequeue_buffer(c->stream);
 if (!b) return;
 struct spa_buffer *sb = b->buffer;
 if (!c->received && sb->n_datas > 0) {
  struct spa_data *d = &sb->datas[0];
  if (d->data && d->chunk && d->chunk->stride > 0 &&
      c->width > 0 && c->height > 0 &&
      (uint32_t)d->chunk->stride >= c->width * 4u) {
   size_t stride = (size_t)d->chunk->stride;
   if (stride <= PA_LIMIT / c->height) {
    size_t len = stride * c->height;
    size_t offset = d->chunk->offset;
    if (offset <= d->maxsize && len <= d->maxsize - offset &&
        len <= d->chunk->size) {
     c->pixels = malloc(len);
     if (c->pixels) {
      memcpy(c->pixels, (unsigned char*)d->data + offset, len);
      c->size = len;
      c->stride = (uint32_t)stride;
      c->received = 1;
     }
    }
   }
  }
 }
 pw_stream_queue_buffer(c->stream,b);
}
static void pa_param_changed(void *userdata, uint32_t id, const struct spa_pod *param) {
 struct pa_capture *c = userdata;
 if (id != SPA_PARAM_Format || !param) return;
 struct spa_video_info_raw info;
 if (spa_format_video_raw_parse(param,&info) < 0) return;
 if (info.format != SPA_VIDEO_FORMAT_BGRA || info.size.width > 8192 ||
     info.size.height > 8192) return;
 c->width = info.size.width;
 c->height = info.size.height;
}
static const struct pw_stream_events pa_events = {
 PW_VERSION_STREAM_EVENTS,
 .process = pa_process,
 .param_changed = pa_param_changed,
};
static struct pa_capture *pa_open(int fd, uint32_t node) {
 if (!node) return NULL;
 pw_init(NULL,NULL);
 struct pa_capture *c = calloc(1,sizeof(*c));
 if (!c) return NULL;
 c->loop = pw_thread_loop_new("ploosassist-local-capture",NULL);
 if (!c->loop) goto fail;
 c->context = pw_context_new(pw_thread_loop_get_loop(c->loop),NULL,0);
 if (!c->context) goto fail;
 // pw_context_connect_fd takes ownership of fd on success.
 c->core = pw_context_connect_fd(c->context,fd,NULL,0);
 if (!c->core) goto fail;
 c->stream = pw_stream_new(c->core,"PloosAssist local screen capture",pw_properties_new(
  PW_KEY_MEDIA_TYPE,"Video",PW_KEY_MEDIA_CATEGORY,"Capture",NULL));
 if (!c->stream) goto fail;
 pw_stream_add_listener(c->stream,&c->listener,&pa_events,c);
 uint8_t buffer[1024];
 struct spa_pod_builder builder = SPA_POD_BUILDER_INIT(buffer,sizeof(buffer));
 const struct spa_pod *params[1];
 params[0] = spa_pod_builder_add_object(&builder,
  SPA_TYPE_OBJECT_Format,SPA_PARAM_EnumFormat,
  SPA_FORMAT_mediaType, SPA_POD_Id(SPA_MEDIA_TYPE_video),
  SPA_FORMAT_mediaSubtype, SPA_POD_Id(SPA_MEDIA_SUBTYPE_raw),
  SPA_FORMAT_VIDEO_format, SPA_POD_Id(SPA_VIDEO_FORMAT_BGRA));
 if (pw_stream_connect(c->stream,PW_DIRECTION_INPUT,node,
  PW_STREAM_FLAG_AUTOCONNECT | PW_STREAM_FLAG_MAP_BUFFERS,
  params,1) < 0) goto fail;
 if (pw_thread_loop_start(c->loop) < 0) goto fail;
 return c;
fail:
 // On early failure, caller retains fd; pa_close is not used here.
 if (c->stream) pw_stream_destroy(c->stream);
 if (c->core) pw_core_disconnect(c->core);
 if (c->context) pw_context_destroy(c->context);
 if (c->loop) pw_thread_loop_destroy(c->loop);
 free(c);
 return NULL;
}
static int pa_poll(struct pa_capture *c) {return c->received;}
static void pa_close(struct pa_capture *c) {
 if (!c) return;
 if (c->loop) pw_thread_loop_stop(c->loop);
 if (c->stream) pw_stream_destroy(c->stream);
 if (c->core) pw_core_disconnect(c->core);
 if (c->context) pw_context_destroy(c->context);
 if (c->loop) pw_thread_loop_destroy(c->loop);
 if (c->pixels) {memset(c->pixels,0,c->size);free(c->pixels);}
 free(c);
}
*/
import "C"

import (
 "context"
 "errors"
 "os"
 "syscall"
 "time"
 "unsafe"

 "github.com/Ploos-AS/PloosAssist/internal/framebuffer"
)

var ErrNoFrame=errors.New("PipeWire did not produce a supported BGRA frame")

// CaptureOne reads a single local BGRA frame from the approved PipeWire node.
// This is an experimental, explicitly opt-in native implementation. The caller
// owns remote and must close it after capture. No network or disk I/O occurs.
func CaptureOne(ctx context.Context,remote *os.File,nodeID uint32,latest *framebuffer.Latest)error{
 if remote==nil||latest==nil||nodeID==0{return errors.New("missing portal remote, node or destination")}
 if err:=ctx.Err();err!=nil{return err}
 fd,err:=syscall.Dup(int(remote.Fd()))
 if err!=nil{return err}
 c:=C.pa_open(C.int(fd),C.uint32_t(nodeID))
 if c==nil{_ =syscall.Close(fd);return errors.New("PipeWire stream initialization failed")}
 defer C.pa_close(c)
 ticker:=time.NewTicker(20*time.Millisecond)
 defer ticker.Stop()
 for{
  select{
  case <-ctx.Done():return ctx.Err()
  case <-ticker.C:
   // This initial adapter is one-shot; the C callback owns the frame
   // until the stream is stopped and its buffers are copied.
   if C.pa_poll(c)==0{continue}
   C.pw_thread_loop_stop(c.loop)
   if c.pixels==nil||c.size==0{return ErrNoFrame}
   pixels:=C.GoBytes(unsafe.Pointer(c.pixels),C.int(c.size))
   return latest.Put(framebuffer.Frame{
    Width:int(c.width),Height:int(c.height),Stride:int(c.stride),Pixels:pixels,
   })
  }
 }
}
