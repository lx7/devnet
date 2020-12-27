package proto

import "github.com/pion/webrtc/v2"

type SessionDescription interface {
	SessionDescription() webrtc.SessionDescription
	IsType(SDP_Type) bool
}

func SDPWithSD(s webrtc.SessionDescription) *SDP {
	return &SDP{
		Type: SDP_Type(s.Type),
		Desc: s.SDP,
	}
}

func (s *SDP) SessionDescription() webrtc.SessionDescription {
	return webrtc.SessionDescription{
		Type: webrtc.SDPType(s.Type),
		SDP:  s.Desc,
	}
}

func (s *SDP) IsType(t SDP_Type) bool {
	return s.Type == t
}

func PayloadWithSD(s webrtc.SessionDescription) *Frame_Sdp {
	return &Frame_Sdp{&SDP{
		Type: SDP_Type(s.Type),
		Desc: s.SDP,
	}}
}

func (f *Frame_Sdp) SessionDescription() webrtc.SessionDescription {
	return webrtc.SessionDescription{
		Type: webrtc.SDPType(f.Sdp.Type),
		SDP:  f.Sdp.Desc,
	}
}

func (f *Frame_Sdp) IsType(t SDP_Type) bool {
	return f.Sdp.Type == t
}
