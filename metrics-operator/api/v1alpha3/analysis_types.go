/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha3

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AnalysisSpec defines the desired state of Analysis
type AnalysisSpec struct {
	// Timeframe specifies the range for the corresponding query in the AnalysisValueTemplate. Please note that either
	// a combination of 'from' and 'to' or the 'recent' property may be set. If neither is set, the Analysis can
	// not be added to the cluster.
	Timeframe `json:"timeframe"`
	// Args corresponds to a map of key/value pairs that can be used to substitute placeholders in the AnalysisValueTemplate query. i.e. for args foo:bar the query could be "query:percentile(95)?scope=tag(my_foo_label:{{.foo}})".
	Args map[string]string `json:"args,omitempty"`
	// AnalysisDefinition refers to the AnalysisDefinition, a CRD that stores the AnalysisValuesTemplates
	AnalysisDefinition ObjectReference `json:"analysisDefinition"`
}

// ProviderResult stores reference of already collected provider query associated to its objective template
type ProviderResult struct {
	// Objective store reference to corresponding objective template
	Objective ObjectReference `json:"objectiveReference,omitempty"`
	// Query represents the executed query
	Query string `json:"query,omitempty"`
	// Value is the value the provider returned
	Value string `json:"value,omitempty"`
	// ErrMsg stores any possible error at retrieval time
	ErrMsg string `json:"errMsg,omitempty"`
}

// AnalysisStatus stores the status of the overall analysis returns also pass or warnings
type AnalysisStatus struct {
	// Timeframe describes the time frame which is evaluated by the Analysis
	Timeframe Timeframe `json:"timeframe"`
	// Raw contains the raw result of the SLO computation
	Raw string `json:"raw,omitempty"`
	// Pass returns whether the SLO is satisfied
	Pass bool `json:"pass,omitempty"`
	// Warning returns whether the analysis returned a warning
	Warning bool `json:"warning,omitempty"`
	// State describes the current state of the Analysis (Pending/Progressing/Completed)
	State AnalysisState `json:"state"`
	// StoredValues contains all analysis values that have already been retrieved successfully
	StoredValues map[string]ProviderResult `json:"storedValues,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="AnalysisDefinition",type=string,JSONPath=.spec.analysisDefinition.name
//+kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Warning",type=string,JSONPath=`.status.warning`
//+kubebuilder:printcolumn:name="Pass",type=string,JSONPath=`.status.pass`

// Analysis is the Schema for the analyses API
type Analysis struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AnalysisSpec   `json:"spec,omitempty"`
	Status AnalysisStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AnalysisList contains a list of Analysis
type AnalysisList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Analysis `json:"items"`
}

type Timeframe struct {
	// From is the time of start for the query. This field follows RFC3339 time format
	From metav1.Time `json:"from,omitempty"`
	// To is the time of end for the query. This field follows RFC3339 time format
	To metav1.Time `json:"to,omitempty"`
	// Recent describes a recent timeframe using a duration string. E.g. Setting this to '5m' provides an Analysis
	// for the last five minutes
	// +optional
	// +kubebuilder:validation:Pattern="^0|([0-9]+(\\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$"
	// +kubebuilder:validation:Type:=string
	// +optional
	Recent metav1.Duration `json:"recent,omitempty"`
}

// GetFrom returns the 'from' timestamp from the status of the Analysis.
// This function has been added to provide a clear way of retrieving the correct timestamp
// to use, which is the one from the Status.
func (a *Analysis) GetFrom() time.Time {
	return a.Status.Timeframe.GetFrom()
}

// GetTo returns the 'from' timestamp from the status of the Analysis.
// This function has been added to provide a clear way of retrieving the correct timestamp
// to use, which is the one from the Status.
func (a *Analysis) GetTo() time.Time {
	return a.Status.Timeframe.GetTo()
}

func (a *Analysis) EnsureTimeframeIsSet() {
	// make sure the correct time frame is set in the status - once an Analysis with a duration string specifying the
	// time frame is triggered, the time frame derived from that duration should stay the same and not shift over the course
	// of multiple reconciliation loops
	if a.Status.Timeframe.From.IsZero() || a.Status.Timeframe.To.IsZero() {
		a.Status.Timeframe.From = metav1.Time{
			Time: a.Spec.GetFrom(),
		}
		a.Status.Timeframe.To = metav1.Time{
			Time: a.Spec.GetTo(),
		}
	}
}

func (t *Timeframe) GetFrom() time.Time {
	if t.Recent.Duration > 0 {
		return time.Now().UTC().Add(-t.Recent.Duration)
	}
	return t.From.Time
}

func (t *Timeframe) GetTo() time.Time {
	if t.Recent.Duration > 0 {
		return time.Now().UTC()
	}
	return t.To.Time
}

func init() {
	SchemeBuilder.Register(&Analysis{}, &AnalysisList{})
}
