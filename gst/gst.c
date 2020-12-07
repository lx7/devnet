#include "gst.h"

#include <gst/app/gstappsrc.h>

typedef struct _CustomData {
  GstElement *pipeline;           
} CustomData;

static guintptr video_window_handle = 0;

static GstBusSyncReply 
bus_sync_handler (GstBus * bus, GstMessage * message, gpointer user_data) {
    if (!gst_is_video_overlay_prepare_window_handle_message (message))
    return GST_BUS_PASS;
    
    if (video_window_handle != 0) {
        GstVideoOverlay *overlay;

        overlay = GST_VIDEO_OVERLAY (GST_MESSAGE_SRC (message));
        gst_video_overlay_set_window_handle (overlay, video_window_handle);
    } else {
        g_warning ("video_window_handle should be defined");
    }

    gst_message_unref (message);
    return GST_BUS_DROP;
}

static void error_cb (GstBus *bus, GstMessage *msg, CustomData *data) {
  GError *err;
  gchar *debug_info;

  gst_message_parse_error (msg, &err, &debug_info);
  g_printerr ("error from element %s: %s\n", GST_OBJECT_NAME (msg->src), err->message);
  g_printerr ("debug: %s\n", debug_info ? debug_info : "none");
  g_clear_error (&err);
  g_free (debug_info);

  gst_element_set_state (data->pipeline, GST_STATE_READY);
}

static void eos_cb (GstBus *bus, GstMessage *msg, CustomData *data) {
    g_print ("END OF STREAM\n");
    gst_element_set_state (data->pipeline, GST_STATE_READY);
}

GstElement *gs_new_pipeline(char *description) {
    gst_init(NULL, NULL);
    GError *error = NULL;
    CustomData data;
    data.pipeline = gst_parse_launch (description, &error);
    
    GstBus *bus = gst_pipeline_get_bus (GST_PIPELINE (data.pipeline));
    
    g_signal_connect (G_OBJECT (bus), "message::error", (GCallback)error_cb, &data);
    g_signal_connect (G_OBJECT (bus), "message::eos", (GCallback)eos_cb, &data);

    gst_bus_set_sync_handler (bus, (GstBusSyncHandler)bus_sync_handler, NULL, NULL);
    gst_object_unref (bus);
    
    return data.pipeline;
}


void gs_pipeline_start (GstElement *pipeline) {
    gst_element_set_state (pipeline, GST_STATE_PLAYING);
}

void gs_pipeline_stop (GstElement *pipeline) { 
    gst_element_set_state (pipeline, GST_STATE_NULL); 
}

void gs_pipeline_set_overlay_handle (GstElement *pipeline, GdkWindow *window) {
    guintptr window_handle;

    if (!gdk_window_ensure_native (window))
        g_error ("Couldn't create native window needed for GstVideoOverlay!");

    /* Retrieve window handler from GDK */
#if defined (GDK_WINDOWING_WIN32)
    video_window_handle = (guintptr)GDK_WINDOW_HWND (window);
#elif defined (GDK_WINDOWING_QUARTZ)
    video_window_handle = gdk_quartz_window_get_nsview (window);
#elif defined (GDK_WINDOWING_X11)
    video_window_handle = GDK_WINDOW_XID (window);
#endif
}

/* GDK helper functions */
GdkWindow *toGdkWindow (guintptr p) {
	return (GDK_WINDOW (p));
}

/* Unit test helper functions */
GMainLoop *test_g_main_loop = NULL;

void test_start_main_loop (void) {
    test_g_main_loop = g_main_loop_new (NULL, FALSE);
    g_main_loop_run (test_g_main_loop);
}

void test_stop_main_loop (void) {
    g_main_loop_quit (test_g_main_loop);
}

