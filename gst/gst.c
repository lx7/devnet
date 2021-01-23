#include "gst.h"

#include <gst/app/gstappsrc.h>

typedef struct _PipelineData {
    int              id;
    GstElement      *pipeline;           
    GstVideoOverlay *overlay;
    GstState         state;
    GtkWidget       *widget;
} PipelineData;

static guintptr video_window_handle[512];

static GstBusSyncReply 
bus_sync_handler (GstBus * bus, GstMessage * message, PipelineData *data) {
    if (!gst_is_video_overlay_prepare_window_handle_message (message))
        return GST_BUS_PASS;
    
    if (data->widget != NULL) {
        guintptr handle;
    
        GdkWindow *window;
        window = gtk_widget_get_parent_window(data->widget);

        if (!gdk_window_ensure_native (window))
            g_error ("failed to get native window for GstVideoOverlay");

        /* Retrieve window handler from GDK */
#if defined (GDK_WINDOWING_WIN32)
        handle = (guintptr)GDK_WINDOW_HWND (window);
#elif defined (GDK_WINDOWING_QUARTZ)
        handle = gdk_quartz_window_get_nsview (window);
#elif defined (GDK_WINDOWING_X11)
        handle = GDK_WINDOW_XID (window);
#endif

        data->overlay = GST_VIDEO_OVERLAY (GST_MESSAGE_SRC (message));
        gst_video_overlay_set_window_handle (data->overlay, handle);

        GtkAllocation al;
        gtk_widget_get_allocation (data->widget, &al);
        
        gst_video_overlay_set_render_rectangle(data->overlay, al.x, al.y, al.width, al.height);
    }

    gst_message_unref (message);
    return GST_BUS_DROP;
}

static gboolean draw_cb (GtkWidget *widget, cairo_t *cr, PipelineData *data) {
    GtkAllocation al;
    gtk_widget_get_allocation (widget, &al);
     
    if (data->overlay != NULL) {
        if (data->state <= GST_STATE_PAUSED) {
            gst_video_overlay_set_render_rectangle (data->overlay, 0, 0, 1, 1);
        } else {
            gst_video_overlay_set_render_rectangle (data->overlay, al.x, al.y, al.width, al.height);
            gst_video_overlay_expose(data->overlay);
        }
    }
    
    if (data->state <= GST_STATE_PAUSED) {
        /*
        gst_video_overlay_set_render_rectangle (data->overlay, 0, 0, 1, 1);
        cairo_set_source_rgb (cr, 0, 0, 0);
        cairo_rectangle (cr, 0, 0, al.width, al.height);
        cairo_fill (cr);
        */
    }

    return FALSE;
}

static void state_cb (GstBus *bus, GstMessage *msg, PipelineData *data) {
    GstState old, new, pending;
    gst_message_parse_state_changed (msg, &old, &new, &pending);
       
    if (GST_MESSAGE_SRC (msg) == GST_OBJECT (data->pipeline)) {
        g_print ("pipeline %i: element %s changed state from %s to %s.\n",
            data->id,
            GST_OBJECT_NAME (msg->src),
            gst_element_state_get_name (old),
            gst_element_state_get_name (new));

        data->state = new;

        if (data->widget != NULL) {
            gtk_widget_queue_draw (data->widget);
        }
    }
}

static void error_cb (GstBus *bus, GstMessage *msg, PipelineData *data) {
    GError *err;
    gchar *debug_info;
    char error_info[512];

    gst_message_parse_error (msg, &err, &debug_info);
    sprintf (error_info,
        "error from element %s: %s\n", 
        GST_OBJECT_NAME (msg->src), 
        err->message
    );
    go_error_cb (data->id, error_info);
    go_error_cb (data->id, debug_info);
    g_clear_error (&err);
    g_free (debug_info);

    gst_element_set_state (data->pipeline, GST_STATE_READY);
}

static void eos_cb (GstBus *bus, GstMessage *msg, PipelineData *data) {
    g_print ("END OF STREAM\n");
    gst_element_set_state (data->pipeline, GST_STATE_READY);
}

static GstFlowReturn sample_cb (GstElement *sink, PipelineData *data) {
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
                    data->id, 
                    copy, copy_size, 
                    GST_BUFFER_DURATION (buf)
                );
        }
        gst_sample_unref (sample);
    }

    return GST_FLOW_OK;
}

void print_debug(const gchar *string) {
    go_debug_cb (-1, (char *)string);
}

void print_error(const gchar *string) {
    go_error_cb (-1, (char *)string);
}

PipelineData *gs_new_pipeline (char *description, int id) {
    gst_init(NULL, NULL);
    GError *err = NULL;
   
    g_set_print_handler (print_debug);
    g_set_printerr_handler (print_error);

    PipelineData *data = calloc(1, sizeof (PipelineData));
    data->pipeline = gst_parse_launch (description, &err);
    if (err != NULL) {
        go_error_cb (id, err->message);
    }
    data->id = id;
    
    video_window_handle[id] = 0;
    
    GstBus *bus = gst_pipeline_get_bus (GST_PIPELINE (data->pipeline));
    gst_bus_add_signal_watch (bus);
    g_signal_connect (G_OBJECT (bus), "message::error", (GCallback)error_cb, data);
    g_signal_connect (G_OBJECT (bus), "message::eos", (GCallback)eos_cb, data);
    g_signal_connect (G_OBJECT (bus), "message::state-changed", (GCallback)state_cb, data);
    gst_bus_set_sync_handler (bus, (GstBusSyncHandler)bus_sync_handler, data, NULL);
    gst_object_unref (bus);
   
    GstElement *sink = gst_bin_get_by_name (GST_BIN (data->pipeline), "sink");
    if (sink != NULL) {
        g_object_set (sink, "emit-signals", TRUE, NULL);
        g_signal_connect (sink, "new-sample", (GCallback)sample_cb, data);
        gst_object_unref (sink);
    }
    
    return data;
}

void gs_pipeline_set_overlay_handle (PipelineData *data, GtkWidget *widget) {
    data->widget = widget;
    g_signal_connect (widget, "draw", G_CALLBACK (draw_cb), data);
}

void gs_pipeline_start (PipelineData *data) {
    gst_element_set_state (data->pipeline, GST_STATE_PLAYING);
}

void gs_pipeline_pause (PipelineData *data) { 
    gst_element_set_state (data->pipeline, GST_STATE_PAUSED); 
}

void gs_pipeline_stop (PipelineData *data) { 
    gst_element_set_state (data->pipeline, GST_STATE_READY); 
}

void gs_pipeline_destroy (PipelineData *data) { 
    gst_element_set_state (data->pipeline, GST_STATE_NULL); 
    gst_object_unref(data->pipeline);
    free(data);
}

void gs_pipeline_appsrc_push (PipelineData *data, void *buf, int len) {
    GstElement *src = gst_bin_get_by_name (GST_BIN (data->pipeline), "src");
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

GtkWidget *to_gtk_widget (guintptr p) {
	return (GTK_WIDGET (p));
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

