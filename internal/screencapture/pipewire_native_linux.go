//go:build linux && cgo && pipewire

package screencapture

/*
#cgo pkg-config: libpipewire-0.3
#include <pipewire/pipewire.h>
#include <spa/param/video/format-utils.h>
#include <spa/param/video/raw-utils.h>
#include <spa/pod/builder.h>
#include <spa/buffer/buffer.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

struct capture {
 struct pw_thread_loop *loop;
 struct pw_context *context;
 struct pw_core *core;
 struct pw_stream *stream;
 struct spa_hook listener;
 struct spa_video_info_raw format;
 unsigned char *pixels;
 size_t capacity;
 size_t length;
 int width, height, stride;
 int frame_ready;
 int failed;
};

static void on_param(void *data, uint32_t id, const struct spa_pod *param) {
 struct capture *c=data;
 if(id!=SPA_PARAM_Format || !param) return;
 struct spa_video_info info={0};
 if(spa_format_parse(param,&info.media_type,&info.media_subtype)<0 ||
    info.media_type!=SPA_MEDIA_TYPE_video ||
    info.media_subtype!=SPA_MEDIA_SUBTYPE_raw ||
    spa_format_video_raw_parse(param,&c->format)<0 ||
    c->format.format!=SPA_VIDEO_FORMAT_BGRA) {
   c->failed=1; return;
 }
 c->width=(int)c->format.size.width;
 c->height=(int)c->format.size.height;
 if(c->width<=0 || c->height<=0 || c->width>8192 || c->height>8192) c->failed=1;
}
static void on_process(void *data) {
 struct capture *c=data;
 struct pw_buffer *b=pw_stream_dequeue_buffer(c->stream);
 if(!b)return;
 struct spa_buffer *buf=b->buffer;
 if(!c->failed && !c->frame_ready && c->width>0 && c->height>0 &&
    buf->n_datas>0 && buf->datas[0].data && buf->datas[0].chunk) {
   struct spa_data *d=&buf->datas[0];
   struct spa_chunk *ch=d->chunk;
   int stride=ch->stride;
   if(stride>=c->width*4 && stride>0 &&
      (size_t)stride<=16777216u/(size_t)c->height &&
      (size_t)ch->offset<=d->maxsize &&
      (size_t)stride*(size_t)c->height<=d->maxsize-(size_t)ch->offset &&
      (size_t)stride*(size_t)c->height<=(size_t)ch->size &&
      (size_t)stride*(size_t)c->height<=c->capacity) {
      size_t n=(size_t)stride*(size_t)c->height;
      memcpy(c->pixels,(unsigned char*)d->data+ch->offset,n);
      c->length=n;c->stride=stride;c->frame_ready=1;
   }
 }
 pw_stream_queue_buffer(c->stream,b);
}
static const struct pw_stream_events events={
 PW_VERSION_STREAM_EVENTS,
 .param_changed=on_param,
 .process=on_process
};
static struct capture *capture_new(int fd,uint32_t node) {
 struct capture *c=calloc(1,sizeof(*c));
 if(!c)return NULL;
 c->capacity=16u*1024u*1024u;
 c->pixels=malloc(c->capacity);
 if(!c->pixels)goto fail;
 pw_init(NULL,NULL);
 c->loop=pw_thread_loop_new("ploosassist-capture",NULL);
 if(!c->loop)goto fail;
 c->context=pw_context_new(pw_thread_loop_get_loop(c->loop),NULL,0);
 if(!c->context)goto fail;
 // pw_context_connect_fd takes ownership of the duplicate.
 int owned=dup(fd);
 if(owned<0)goto fail;
 c->core=pw_context_connect_fd(c->context,owned,NULL,0);
 if(!c->core){close(owned);goto fail;}
 c->stream=pw_stream_new(c->core,"ploosassist-local-capture",
  pw_properties_new(PW_KEY_MEDIA_TYPE,"Video",PW_KEY_MEDIA_CATEGORY,"Capture",NULL));
 if(!c->stream)goto fail;
 pw_stream_add_listener(c->stream,&c->listener,&events,c);
 uint8_t buffer[1024];
 struct spa_pod_builder builder=SPA_POD_BUILDER_INIT(buffer,sizeof(buffer));
 const struct spa_pod *params[1];
 params[0]=(const struct spa_pod*)spa_pod_builder_add_object(&builder,
   SPA_TYPE_OBJECT_Format,SPA_PARAM_EnumFormat,
   SPA_FORMAT_mediaType,SPA_POD_Id(SPA_MEDIA_TYPE_video),
   SPA_FORMAT_mediaSubtype,SPA_POD_Id(SPA_MEDIA_SUBTYPE_raw),
   SPA_FORMAT_VIDEO_format,SPA_POD_Id(SPA_VIDEO_FORMAT_BGRA));
 if(pw_stream_connect(c->stream,PW_DIRECTION_INPUT,node,
   PW_STREAM_FLAG_AUTOCONNECT|PW_STREAM_FLAG_MAP_BUFFERS,params,1)<0)goto fail;
 if(pw_thread_loop_start(c->loop)<0)goto fail;
 return c;
fail:
 if(c->stream)pw_stream_destroy(c->stream);
 if(c->core)pw_core_disconnect(c->core);
 if(c->context)pw_context_destroy(c->context);
 if(c->loop)pw_thread_loop_destroy(c->loop);
 free(c->pixels);free(c);return NULL;
}
static int capture_poll(struct capture *c,void *dst,size_t max,int *w,int *h,int *stride) {
 pw_thread_loop_lock(c->loop);
 int result=0;
 if(c->failed)result=-1;
 else if(c->frame_ready && c->length<=max) {
   memcpy(dst,c->pixels,c->length);
   *w=c->width;*h=c->height;*stride=c->stride;
   result=(int)c->length;
   c->frame_ready=0;
 }
 pw_thread_loop_unlock(c->loop);
 return result;
}
static void capture_free(struct capture *c) {
 if(!c)return;
 if(c->loop)pw_thread_loop_stop(c->loop);
 if(c->stream)pw_stream_destroy(c->stream);
 if(c->core)pw_core_disconnect(c->core);
 if(c->context)pw_context_destroy(c->context);
 if(c->loop)pw_thread_loop_destroy(c->loop);
 if(c->pixels){memset(c->pixels,0,c->capacity);free(c->pixels);}
 free(c);
}
*/
import "C"

import (
 "context"
 "errors"
 "os"
 "time"
 "unsafe"

 "github.com/Ploos-AS/PloosAssist/internal/framebuffer"
)

// NativePipeWireConsumer uses the portal FD and approved node ID only.
// Build explicitly with -tags pipewire and libpipewire-0.3 development headers.
type NativePipeWireConsumer struct{}

func (NativePipeWireConsumer) Run(ctx context.Context,fd *os.File,node uint32,dst *framebuffer.Latest)error{
 if fd==nil||node==0||dst==nil{return ErrNotAuthorized}
 if err:=ctx.Err();err!=nil{return err}
 c:=C.capture_new(C.int(fd.Fd()),C.uint32_t(node))
 if c==nil{return errors.New("native PipeWire stream initialization failed")}
 defer C.capture_free(c)
 pixels:=make([]byte,framebuffer.MaxFrameBytes)
 defer clear(pixels)
 ticker:=time.NewTicker(20*time.Millisecond)
 defer ticker.Stop()
 for{
  select{
  case <-ctx.Done():return ctx.Err()
  case <-ticker.C:
   var w,h,stride C.int
   n:=int(C.capture_poll(c,unsafe.Pointer(&pixels[0]),C.size_t(len(pixels)),&w,&h,&stride))
   if n<0{return errors.New("PipeWire negotiated an unsupported video format")}
   if n>0{
    if err:=dst.Put(framebuffer.Frame{
     Width:int(w),Height:int(h),Stride:int(stride),Pixels:pixels[:n],
    });err!=nil{return err}
   }
  }
 }
}
