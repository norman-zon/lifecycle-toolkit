package keptnappversion

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	lfcv1alpha3 "github.com/keptn/lifecycle-toolkit/lifecycle-operator/apis/lifecycle/v1alpha3"
	apicommon "github.com/keptn/lifecycle-toolkit/lifecycle-operator/apis/lifecycle/v1alpha3/common"
	lfcv1alpha4 "github.com/keptn/lifecycle-toolkit/lifecycle-operator/apis/lifecycle/v1alpha4"
	"github.com/keptn/lifecycle-toolkit/lifecycle-operator/controllers/common/evaluation"
	evalfake "github.com/keptn/lifecycle-toolkit/lifecycle-operator/controllers/common/evaluation/fake"
	"github.com/keptn/lifecycle-toolkit/lifecycle-operator/controllers/common/eventsender"
	telemetryfake "github.com/keptn/lifecycle-toolkit/lifecycle-operator/controllers/common/telemetry/fake"
	"github.com/keptn/lifecycle-toolkit/lifecycle-operator/controllers/common/testcommon"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type contextID string

const CONTEXTID contextID = "start"

// this test checks if the chain of reconcile events is correct
func TestKeptnAppVersionReconciler_reconcile(t *testing.T) {

	pendingStatus := lfcv1alpha3.KeptnAppVersionStatus{
		CurrentPhase:                   "",
		Status:                         apicommon.StatePending,
		PreDeploymentStatus:            apicommon.StatePending,
		PreDeploymentEvaluationStatus:  apicommon.StatePending,
		WorkloadOverallStatus:          apicommon.StatePending,
		PostDeploymentStatus:           apicommon.StatePending,
		PostDeploymentEvaluationStatus: apicommon.StatePending,
	}

	app := testcommon.ReturnAppVersion("default", "myappversion", "1.0.0", nil, pendingStatus)

	r, eventChannel, _ := setupReconciler(app)

	tests := []struct {
		name    string
		req     ctrl.Request
		wantErr error
		events  []string // check correct events are generated
	}{
		{
			name: "new appVersion with no workload nor evaluation should finish",
			req: ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: "default",
					Name:      "myappversion-1.0.0",
				},
			},
			wantErr: nil,
			events: []string{
				`AppPreDeployTasksStarted`,
				`AppPreDeployTasksFinished`,
				`AppPreDeployEvaluationsStarted`,
				`AppPreDeployEvaluationsFinished`,
				`AppDeployStarted`,
				`AppDeployFinished`,
				`AppPostDeployTasksStarted`,
				`AppPostDeployTasksFinished`,
				`AppPostDeployEvaluationsStarted`,
				`AppPostDeployEvaluationsFinished`,
			},
		},
		{
			name: "notfound should not return error nor event",
			req: ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: "default",
					Name:      "mynotthereapp",
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			_, err := r.Reconcile(context.WithValue(context.TODO(), CONTEXTID, tt.req.Name), tt.req)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("Reconcile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.events != nil {
				for _, e := range tt.events {
					event := <-eventChannel
					require.Equal(t, strings.Contains(event, tt.req.Name), true, "wrong appversion")
					require.Equal(t, strings.Contains(event, tt.req.Namespace), true, "wrong namespace")
					require.Equal(t, strings.Contains(event, e), true, fmt.Sprintf("no %s found in %s", e, event))
				}

			}
		})

	}

}

func TestKeptnAppVersionReconciler_ReconcileFailed(t *testing.T) {

	status := lfcv1alpha3.KeptnAppVersionStatus{
		CurrentPhase:        apicommon.PhaseAppPreDeployment.ShortName,
		Status:              apicommon.StateProgressing,
		PreDeploymentStatus: apicommon.StateProgressing,
		PreDeploymentTaskStatus: []lfcv1alpha3.ItemStatus{
			{
				Name:           "pre-task",
				DefinitionName: "task",
				Status:         apicommon.StateFailed,
			},
		},
		PreDeploymentEvaluationStatus:  apicommon.StatePending,
		WorkloadOverallStatus:          apicommon.StatePending,
		PostDeploymentStatus:           apicommon.StatePending,
		PostDeploymentEvaluationStatus: apicommon.StatePending,
	}

	appVersionName := fmt.Sprintf("%s-%s", "myapp", "1.0.0")
	app := &lfcv1alpha3.KeptnAppVersion{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:       appVersionName,
			Namespace:  "default",
			Generation: 1,
		},
		Spec: lfcv1alpha3.KeptnAppVersionSpec{
			KeptnAppSpec: lfcv1alpha3.KeptnAppSpec{
				Version: "1.0.0",
				PreDeploymentTasks: []string{
					"task",
				},
			},
			AppName: "myapp",
			TraceId: map[string]string{
				"traceparent": "parent-trace",
			},
		},
		Status: status,
	}
	r, eventChannel, _ := setupReconciler(app)

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "default",
			Name:      "myapp-1.0.0",
		},
	}

	result, err := r.Reconcile(context.WithValue(context.TODO(), CONTEXTID, req.Name), req)
	require.Nil(t, err)

	expectedEvents := []string{
		"AppPreDeployTasksFailed",
	}

	for _, e := range expectedEvents {
		event := <-eventChannel
		require.Equal(t, strings.Contains(event, req.Name), true, "wrong appversion")
		require.Equal(t, strings.Contains(event, req.Namespace), true, "wrong namespace")
		require.Equal(t, strings.Contains(event, e), true, fmt.Sprintf("no %s found in %s", e, event))
	}

	require.Nil(t, err)

	// do not requeue since we reached completion
	require.False(t, result.Requeue)
}

func TestKeptnAppVersionReconciler_ReconcileReachCompletion(t *testing.T) {

	app := testcommon.ReturnAppVersion("default", "myfinishedapp", "1.0.0", nil, createFinishedAppVersionStatus())
	r, eventChannel, _ := setupReconciler(app)
	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "default",
			Name:      "myfinishedapp-1.0.0",
		},
	}

	result, err := r.Reconcile(context.WithValue(context.TODO(), CONTEXTID, req.Name), req)
	require.Nil(t, err)

	expectedEvents := []string{
		"CompletedFinished",
	}

	for _, e := range expectedEvents {
		event := <-eventChannel
		require.Equal(t, strings.Contains(event, req.Name), true, "wrong appversion")
		require.Equal(t, strings.Contains(event, req.Namespace), true, "wrong namespace")
		require.Equal(t, strings.Contains(event, e), true, fmt.Sprintf("no %s found in %s", e, event))
	}

	require.Nil(t, err)

	// do not requeue since we reached completion
	require.False(t, result.Requeue)
}

func createFinishedAppVersionStatus() lfcv1alpha3.KeptnAppVersionStatus {
	return lfcv1alpha3.KeptnAppVersionStatus{
		CurrentPhase:                       apicommon.PhaseCompleted.ShortName,
		PreDeploymentStatus:                apicommon.StateSucceeded,
		PostDeploymentStatus:               apicommon.StateSucceeded,
		PreDeploymentEvaluationStatus:      apicommon.StateSucceeded,
		PostDeploymentEvaluationStatus:     apicommon.StateSucceeded,
		PreDeploymentTaskStatus:            []lfcv1alpha3.ItemStatus{{Status: apicommon.StateSucceeded}},
		PostDeploymentTaskStatus:           []lfcv1alpha3.ItemStatus{{Status: apicommon.StateSucceeded}},
		PreDeploymentEvaluationTaskStatus:  []lfcv1alpha3.ItemStatus{{Status: apicommon.StateSucceeded}},
		PostDeploymentEvaluationTaskStatus: []lfcv1alpha3.ItemStatus{{Status: apicommon.StateSucceeded}},
		WorkloadOverallStatus:              apicommon.StateSucceeded,
		WorkloadStatus:                     []lfcv1alpha3.WorkloadStatus{{Status: apicommon.StateSucceeded}},
		Status:                             apicommon.StateSucceeded,
	}
}

func setupReconcilerWithMeters() *KeptnAppVersionReconciler {
	// setup logger
	opts := zap.Options{
		Development: true,
	}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// fake a tracer
	tr := &telemetryfake.ITracerMock{StartFunc: func(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
		return ctx, trace.SpanFromContext(ctx)
	}}

	tf := &telemetryfake.TracerFactoryMock{GetTracerFunc: func(name string) trace.Tracer {
		return tr
	}}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	r := &KeptnAppVersionReconciler{
		Log:           ctrl.Log.WithName("test-appVersionController"),
		TracerFactory: tf,
		Meters:        testcommon.InitAppMeters(),
	}
	return r
}

func setupReconciler(objs ...client.Object) (*KeptnAppVersionReconciler, chan string, *telemetryfake.ISpanHandlerMock) {
	// setup logger
	opts := zap.Options{
		Development: true,
	}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// fake a tracer
	tr := &telemetryfake.ITracerMock{StartFunc: func(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
		return ctx, trace.SpanFromContext(ctx)
	}}

	tf := &telemetryfake.TracerFactoryMock{GetTracerFunc: func(name string) trace.Tracer {
		return tr
	}}

	// fake span handler

	spanRecorder := &telemetryfake.ISpanHandlerMock{
		GetSpanFunc: func(ctx context.Context, tracer trace.Tracer, reconcileObject client.Object, phase string) (context.Context, trace.Span, error) {
			return ctx, trace.SpanFromContext(ctx), nil
		},
		UnbindSpanFunc: func(reconcileObject client.Object, phase string) error { return nil },
	}

	workloadVersionIndexer := func(obj client.Object) []string {
		workloadVersion, _ := obj.(*lfcv1alpha4.KeptnWorkloadVersion)
		return []string{workloadVersion.Spec.AppName}
	}

	testcommon.SetupSchemes()
	fakeClient := fake.NewClientBuilder().WithObjects(objs...).WithStatusSubresource(objs...).WithScheme(scheme.Scheme).WithObjects().WithIndex(&lfcv1alpha4.KeptnWorkloadVersion{}, "spec.app", workloadVersionIndexer).Build()

	recorder := record.NewFakeRecorder(100)
	r := &KeptnAppVersionReconciler{
		Client:        fakeClient,
		Scheme:        scheme.Scheme,
		EventSender:   eventsender.NewK8sSender(recorder),
		Log:           ctrl.Log.WithName("test-appVersionController"),
		TracerFactory: tf,
		SpanHandler:   spanRecorder,
		Meters:        testcommon.InitAppMeters(),
		EvaluationHandler: &evalfake.MockEvaluationHandler{
			ReconcileEvaluationsFunc: func(ctx context.Context, phaseCtx context.Context, reconcileObject client.Object, evaluationCreateAttributes evaluation.CreateEvaluationAttributes) ([]lfcv1alpha3.ItemStatus, apicommon.StatusSummary, error) {
				return []lfcv1alpha3.ItemStatus{}, apicommon.StatusSummary{}, nil
			},
		},
	}
	return r, recorder.Events, spanRecorder
}

func TestKeptnApVersionReconciler_setupSpansContexts(t *testing.T) {

	r := setupReconcilerWithMeters()
	type args struct {
		ctx        context.Context
		appVersion *lfcv1alpha3.KeptnAppVersion
	}
	tests := []struct {
		name    string
		args    args
		baseCtx context.Context
	}{
		{
			name: "Current trace ctx should be != than app trace context",
			args: args{
				ctx: context.WithValue(context.TODO(), CONTEXTID, 1),
				appVersion: &lfcv1alpha3.KeptnAppVersion{
					Spec: lfcv1alpha3.KeptnAppVersionSpec{TraceId: map[string]string{
						"traceparent": "00-52527d549a7b33653017ce960be09dfc-a38a5a8d179a88b5-01",
					}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, endFunc := r.setupSpansContexts(tt.args.ctx, tt.args.appVersion)
			require.NotNil(t, ctx)
			require.NotNil(t, endFunc)

		})
	}
}
