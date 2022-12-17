package controllers

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type mockEventRecorder struct{}

func (m mockEventRecorder) Event(object runtime.Object, eventtype, reason, message string) {}
func (m mockEventRecorder) Eventf(object runtime.Object, eventtype, reason, message string, args ...interface{}) {
}
func (m mockEventRecorder) AnnotatedEventf(object runtime.Object, annotations map[string]string, eventtype, reason, messageFmt string, args ...interface{}) {
}

func TestProcessRequeue(t *testing.T) {
	fakeReconcile := struct{}{}
	fakeObj := &corev1.ConfigMap{}
	fakeEventRecorder := &mockEventRecorder{}
	fakeLogger := log.WithName("fake logger")

	t.Run("should return err == nil when requeue.curError equal nil", func(t *testing.T) {
		requeue := requeue{
			message: "test",
		}
		result, err := processRequeue(&requeue, fakeReconcile, fakeObj, fakeEventRecorder, fakeLogger)
		assert.True(t, result.Requeue)
		assert.Zero(t, result.RequeueAfter)
		assert.Nil(t, err)
	})

	t.Run("should return error", func(t *testing.T) {
		requeue := requeue{
			curError: errors.New("fake"),
		}
		result, err := processRequeue(&requeue, fakeReconcile, fakeObj, fakeEventRecorder, fakeLogger)
		assert.False(t, result.Requeue)
		assert.Zero(t, result.RequeueAfter)
		assert.Error(t, err, "fake")
	})

	t.Run("should return err == nil, when requeue.curError is Conflict", func(t *testing.T) {
		k8sErr := &k8sErrors.StatusError{
			ErrStatus: metav1.Status{
				Reason: metav1.StatusReasonConflict,
			},
		}
		requeue := requeue{
			curError: k8sErr,
		}
		result, err := processRequeue(&requeue, fakeReconcile, fakeObj, fakeEventRecorder, fakeLogger)
		assert.True(t, result.Requeue)
		assert.Nil(t, err)
	})
}
