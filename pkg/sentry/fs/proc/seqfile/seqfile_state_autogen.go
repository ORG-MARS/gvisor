// automatically generated by stateify.

package seqfile

import (
	"gvisor.dev/gvisor/pkg/state"
)

func (x *SeqData) beforeSave() {}
func (x *SeqData) save(m state.Map) {
	x.beforeSave()
	m.Save("Buf", &x.Buf)
	m.Save("Handle", &x.Handle)
}

func (x *SeqData) afterLoad() {}
func (x *SeqData) load(m state.Map) {
	m.Load("Buf", &x.Buf)
	m.Load("Handle", &x.Handle)
}

func (x *SeqFile) beforeSave() {}
func (x *SeqFile) save(m state.Map) {
	x.beforeSave()
	m.Save("InodeSimpleExtendedAttributes", &x.InodeSimpleExtendedAttributes)
	m.Save("InodeSimpleAttributes", &x.InodeSimpleAttributes)
	m.Save("SeqSource", &x.SeqSource)
	m.Save("source", &x.source)
	m.Save("generation", &x.generation)
	m.Save("lastRead", &x.lastRead)
}

func (x *SeqFile) afterLoad() {}
func (x *SeqFile) load(m state.Map) {
	m.Load("InodeSimpleExtendedAttributes", &x.InodeSimpleExtendedAttributes)
	m.Load("InodeSimpleAttributes", &x.InodeSimpleAttributes)
	m.Load("SeqSource", &x.SeqSource)
	m.Load("source", &x.source)
	m.Load("generation", &x.generation)
	m.Load("lastRead", &x.lastRead)
}

func (x *seqFileOperations) beforeSave() {}
func (x *seqFileOperations) save(m state.Map) {
	x.beforeSave()
	m.Save("seqFile", &x.seqFile)
}

func (x *seqFileOperations) afterLoad() {}
func (x *seqFileOperations) load(m state.Map) {
	m.Load("seqFile", &x.seqFile)
}

func init() {
	state.Register("pkg/sentry/fs/proc/seqfile.SeqData", (*SeqData)(nil), state.Fns{Save: (*SeqData).save, Load: (*SeqData).load})
	state.Register("pkg/sentry/fs/proc/seqfile.SeqFile", (*SeqFile)(nil), state.Fns{Save: (*SeqFile).save, Load: (*SeqFile).load})
	state.Register("pkg/sentry/fs/proc/seqfile.seqFileOperations", (*seqFileOperations)(nil), state.Fns{Save: (*seqFileOperations).save, Load: (*seqFileOperations).load})
}