#ifndef GST_H
#define GST_H

#include <gtk/gtk.h>
#include <glib.h>
#include <gst/gst.h>
#include <gst/video/videooverlay.h>
#include <stdint.h>
#include <stdlib.h>

#include <gdk/gdk.h>
#if defined (GDK_WINDOWING_X11)
#include <gdk/gdkx.h>
#elif defined (GDK_WINDOWING_WIN32)
#include <gdk/gdkwin32.h>
#elif defined (GDK_WINDOWING_QUARTZ)
#include <gdk/gdkquartzwindow.h>
#endif

GstElement *gs_new_pipeline(char *pipeline);
void gs_pipeline_set_overlay_handle (GstElement *pipeline, GdkWindow *window);
void gs_pipeline_start (GstElement *element);
void gs_pipeline_stop (GstElement *element);

/* GDK helper functions */
GdkWindow *toGdkWindow (guintptr p);

/* Unit test helper functions */
void test_start_main_loop(void);
void test_stop_main_loop(void);

#endif
