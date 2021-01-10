package proto

import "github.com/pion/webrtc/v3"

func (i *ICE) ICECandidate() webrtc.ICECandidateInit {
	mli := uint16(i.SdpMLineIndex)
	return webrtc.ICECandidateInit{
		Candidate:        i.Candidate,
		SDPMid:           &i.SdpMid,
		SDPMLineIndex:    &mli,
		UsernameFragment: &i.UsernameFragment,
	}
}

func PayloadWithICECandidate(i webrtc.ICECandidateInit) *Frame_Ice {
	f := &Frame_Ice{&ICE{
		Candidate:     i.Candidate,
		SdpMid:        *i.SDPMid,
		SdpMLineIndex: uint32(*i.SDPMLineIndex),
	}}

	if i.UsernameFragment != nil {
		f.Ice.UsernameFragment = *i.UsernameFragment
	}
	return f
}

func (f *Frame_Ice) ICECandidate() webrtc.ICECandidateInit {
	return f.Ice.ICECandidate()
}
