package proto

import "github.com/pion/webrtc/v3"

func (s *SDP) SessionDescription() webrtc.SessionDescription {
	return webrtc.SessionDescription{
		Type: webrtc.SDPType(s.Type),
		SDP:  s.Desc,
	}
}

func PayloadWithSD(s webrtc.SessionDescription) *Frame_Sdp {
	return &Frame_Sdp{&SDP{
		Type: SDP_Type(s.Type),
		Desc: s.SDP,
	}}
}

func (f *Frame_Sdp) SessionDescription() webrtc.SessionDescription {
	return f.Sdp.SessionDescription()
}
