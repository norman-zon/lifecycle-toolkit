// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package fake

import (
	"context"
	apicommon "github.com/keptn/lifecycle-toolkit/lifecycle-operator/apis/lifecycle/v1alpha3/common"
	"github.com/keptn/lifecycle-toolkit/lifecycle-operator/controllers/common/phase"
	"go.opentelemetry.io/otel/trace"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

// MockHandler is a mock implementation of phase.IHandler.
//
//	func TestSomethingThatUsesIHandler(t *testing.T) {
//
//		// make and configure a mocked phase.IHandler
//		mockedIHandler := &MockHandler{
//			HandlePhaseFunc: func(ctx context.Context, ctxTrace context.Context, tracer trace.Tracer, reconcileObject client.Object, phaseMoqParam apicommon.KeptnPhaseType, reconcilePhase func(phaseCtx context.Context) (apicommon.KeptnState, error)) (phase.PhaseResult, error) {
//				panic("mock out the HandlePhase method")
//			},
//		}
//
//		// use mockedIHandler in code that requires phase.IHandler
//		// and then make assertions.
//
//	}
type MockHandler struct {
	// HandlePhaseFunc mocks the HandlePhase method.
	HandlePhaseFunc func(ctx context.Context, ctxTrace context.Context, tracer trace.Tracer, reconcileObject client.Object, phaseMoqParam apicommon.KeptnPhaseType, reconcilePhase func(phaseCtx context.Context) (apicommon.KeptnState, error)) (phase.PhaseResult, error)

	// calls tracks calls to the methods.
	calls struct {
		// HandlePhase holds details about calls to the HandlePhase method.
		HandlePhase []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// CtxTrace is the ctxTrace argument value.
			CtxTrace context.Context
			// Tracer is the tracer argument value.
			Tracer trace.Tracer
			// ReconcileObject is the reconcileObject argument value.
			ReconcileObject client.Object
			// PhaseMoqParam is the phaseMoqParam argument value.
			PhaseMoqParam apicommon.KeptnPhaseType
			// ReconcilePhase is the reconcilePhase argument value.
			ReconcilePhase func(phaseCtx context.Context) (apicommon.KeptnState, error)
		}
	}
	lockHandlePhase sync.RWMutex
}

// HandlePhase calls HandlePhaseFunc.
func (mock *MockHandler) HandlePhase(ctx context.Context, ctxTrace context.Context, tracer trace.Tracer, reconcileObject client.Object, phaseMoqParam apicommon.KeptnPhaseType, reconcilePhase func(phaseCtx context.Context) (apicommon.KeptnState, error)) (phase.PhaseResult, error) {
	if mock.HandlePhaseFunc == nil {
		panic("MockHandler.HandlePhaseFunc: method is nil but IHandler.HandlePhase was just called")
	}
	callInfo := struct {
		Ctx             context.Context
		CtxTrace        context.Context
		Tracer          trace.Tracer
		ReconcileObject client.Object
		PhaseMoqParam   apicommon.KeptnPhaseType
		ReconcilePhase  func(phaseCtx context.Context) (apicommon.KeptnState, error)
	}{
		Ctx:             ctx,
		CtxTrace:        ctxTrace,
		Tracer:          tracer,
		ReconcileObject: reconcileObject,
		PhaseMoqParam:   phaseMoqParam,
		ReconcilePhase:  reconcilePhase,
	}
	mock.lockHandlePhase.Lock()
	mock.calls.HandlePhase = append(mock.calls.HandlePhase, callInfo)
	mock.lockHandlePhase.Unlock()
	return mock.HandlePhaseFunc(ctx, ctxTrace, tracer, reconcileObject, phaseMoqParam, reconcilePhase)
}

// HandlePhaseCalls gets all the calls that were made to HandlePhase.
// Check the length with:
//
//	len(mockedIHandler.HandlePhaseCalls())
func (mock *MockHandler) HandlePhaseCalls() []struct {
	Ctx             context.Context
	CtxTrace        context.Context
	Tracer          trace.Tracer
	ReconcileObject client.Object
	PhaseMoqParam   apicommon.KeptnPhaseType
	ReconcilePhase  func(phaseCtx context.Context) (apicommon.KeptnState, error)
} {
	var calls []struct {
		Ctx             context.Context
		CtxTrace        context.Context
		Tracer          trace.Tracer
		ReconcileObject client.Object
		PhaseMoqParam   apicommon.KeptnPhaseType
		ReconcilePhase  func(phaseCtx context.Context) (apicommon.KeptnState, error)
	}
	mock.lockHandlePhase.RLock()
	calls = mock.calls.HandlePhase
	mock.lockHandlePhase.RUnlock()
	return calls
}
