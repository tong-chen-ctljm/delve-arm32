package native

import (
	"fmt"

	sys "golang.org/x/sys/unix"

	"github.com/go-delve/delve/pkg/proc"
)

type waitStatus sys.WaitStatus

// osSpecificDetails hold Linux specific
// process details.
type osSpecificDetails struct {
	delayedSignal       int
	running             bool
	setbp               bool
	phantomBreakpointPC uint64
}

func (t *nativeThread) stop() (err error) {
	err = sys.Tgkill(t.dbp.pid, t.ID, sys.SIGSTOP)
	if err != nil {
		err = fmt.Errorf("stop err %s on thread %d", err, t.ID)
		return
	}
	return
}

// Stopped returns whether the thread is stopped at
// the operating system level.
func (t *nativeThread) Stopped() bool {
	state := status(t.ID, t.dbp.os.comm)
	return state == statusTraceStop || state == statusTraceStopT
}

func (t *nativeThread) resume() error {
	sig := t.os.delayedSignal
	t.os.delayedSignal = 0
	return t.resumeWithSig(sig)
}

func (t *nativeThread) resumeWithSig(sig int) (err error) {
	t.os.running = true
	t.dbp.execPtraceFunc(func() { err = ptraceCont(t.ID, sig) })
	return
}

func (t *nativeThread) Blocked() bool {
	regs, err := t.Registers()
	if err != nil {
		return false
	}
	pc := regs.PC()
	fn := t.BinInfo().PCToFunc(pc)
	if fn != nil && ((fn.Name == "runtime.futex") || (fn.Name == "runtime.usleep") || (fn.Name == "runtime.clone")) {
		return true
	}
	return false
}



func (t *nativeThread) WriteMemory(addr uint64, data []byte) (written int, err error) {
	if t.dbp.exited {
		return 0, proc.ErrProcessExited{Pid: t.dbp.pid}
	}
	if len(data) == 0 {
		return
	}
	// ProcessVmWrite can't poke read-only memory like ptrace, so don't
	// even bother for small writes -- likely breakpoints and such.
	if len(data) > sys.SizeofPtr {
		written, _ = processVmWrite(t.ID, uintptr(addr), data)
	}
	if written == 0 {
		t.dbp.execPtraceFunc(func() { written, err = sys.PtracePokeData(t.ID, uintptr(addr), data) })
	}
	return
}

func (t *nativeThread) ReadMemory(data []byte, addr uint64) (n int, err error) {
	if t.dbp.exited {
		return 0, proc.ErrProcessExited{Pid: t.dbp.pid}
	}
	if len(data) == 0 {
		return
	}
	n, _ = processVmRead(t.ID, uintptr(addr), data)
	if n == 0 {
		t.dbp.execPtraceFunc(func() { n, err = sys.PtracePeekData(t.ID, uintptr(addr), data) })
	}
	return
}

// SoftExc returns true if this thread received a software exception during the last resume.
func (t *nativeThread) SoftExc() bool {
	return t.os.setbp
}
