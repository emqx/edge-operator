package mock

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
)

type mockEventRecorder struct {
}

func (m *mockEventRecorder) Event(_ runtime.Object, _, _, _ string) {

}

// Eventf is just like Event, but with Sprintf for the message field.
func (m *mockEventRecorder) Eventf(_ runtime.Object, _, _, _ string, _ ...interface{}) {

}

// AnnotatedEventf is just like eventf, but with annotations attached
func (m *mockEventRecorder) AnnotatedEventf(_ runtime.Object, _ map[string]string, _, _, _ string, _ ...interface{}) {

}

func GetEventRecorderFor(_ string) record.EventRecorder {
	return &mockEventRecorder{}
}
