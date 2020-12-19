#include "gst.h"

#include <gst/app/gstappsrc.h>

typedef struct _CustomData {
    GstElement *pipeline;           
    int pipeline_id;
} CustomData;

static guintptr video_window_handle = 0;

static GstBusSyncReply 
bus_sync_handler (GstBus * bus, GstMessage * message, CustomData *data) {
    if (!gst_is_video_overlay_prepare_window_handle_message (message))
        return GST_BUS_PASS;
    
    if (video_window_handle != 0) {
        GstVideoOverlay *overlay;

        overlay = GST_VIDEO_OVERLAY (GST_MESSAGE_SRC (message));
        gst_video_overlay_set_window_handle (overlay, video_window_handle);
    }

    gst_message_unref (message);
    return GST_BUS_DROP;
}

static void error_cb (GstBus *bus, GstMessage *msg, CustomData *data) {
    GError *err;
    gchar *debug_info;
    gchar *error_info;

    gst_message_parse_error (msg, &err, &debug_info);

    sprintf (error_info,
            "error from element %s: %s\n", 
            GST_OBJECT_NAME (msg->src), 
            err->message
        );
    go_error_cb (data->pipeline_id, error_info);
    go_debug_cb (data->pipeline_id, debug_info);
    g_clear_error (&err);
    g_free (debug_info);
    g_free (error_info);

    gst_element_set_state (data->pipeline, GST_STATE_READY);
}

static void eos_cb (GstBus *bus, GstMessage *msg, CustomData *data) {
    g_print ("END OF STREAM\n");
    gst_element_set_state (data->pipeline, GST_STATE_READY);
}

static GstFlowReturn sample_cb (GstElement *sink, CustomData *data) {
    GstSample *sample = NULL;
    GstBuffer *buf = NULL;
    gpointer copy = NULL;
    gsize copy_size = 0;

    g_signal_emit_by_name (sink, "pull-sample", &sample);
    if (sample) {
        buf = gst_sample_get_buffer (sample);
        if (buf) {
            gst_buffer_extract_dup (
                    buf, 0, gst_buffer_get_size (buf), 
                    &copy, &copy_size
                );
            go_sample_cb (
                    data->pipeline_id, 
                    copy, copy_size, 
                    GST_BUFFER_DURATION (buf)
                );
        }
        gst_sample_unref (sample);
    }

    return GST_FLOW_OK;
}

GstElement *gs_new_pipeline (char *description, int id) {
    gst_init(NULL, NULL);
    GError *err = NULL;

    CustomData *data = calloc(1, sizeof (CustomData));
    data->pipeline = gst_parse_launch (description, &err);
    if (err != NULL) {
        go_error_cb (id, err->message);
        return NULL;
    }
    data->pipeline_id = id;
    
    GstBus *bus = gst_pipeline_get_bus (GST_PIPELINE (data->pipeline));
    g_signal_connect (G_OBJECT (bus), "message::error", (GCallback)error_cb, data);
    g_signal_connect (G_OBJECT (bus), "message::eos", (GCallback)eos_cb, data);
    gst_bus_set_sync_handler (bus, (GstBusSyncHandler)bus_sync_handler, data, NULL);
    gst_object_unref (bus);
  
    GstElement *sink = gst_bin_get_by_name (GST_BIN (data->pipeline), "sink");
    if (sink != NULL) {
        g_object_set (sink, "emit-signals", TRUE, NULL);
        g_signal_connect (sink, "new-sample", (GCallback)sample_cb, data);
        gst_object_unref (sink);
    }
    
    return data->pipeline;
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

void gs_pipeline_start (GstElement *pipeline) {
    gst_element_set_state (pipeline, GST_STATE_PLAYING);
}

void gs_pipeline_stop (GstElement *pipeline) { 
    gst_element_set_state (pipeline, GST_STATE_NULL); 
}

void gs_pipeline_destroy (GstElement *pipeline) { 
    gst_element_set_state (pipeline, GST_STATE_NULL); 
    gst_object_unref(pipeline);
}

void gs_pipeline_appsrc_push (GstElement *pipeline, void *buf, int len) {
    GstElement *src = gst_bin_get_by_name (GST_BIN (pipeline), "src");
    if (src != NULL) {
        gpointer p = g_memdup (buf, len);
        GstBuffer *buf = gst_buffer_new_wrapped (p, len);
        gst_app_src_push_buffer (GST_APP_SRC (src), buf);
        gst_object_unref (src);
    }
}

/* GDK helper functions */
GdkWindow *to_gdk_window (guintptr p) {
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

