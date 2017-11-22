package main

import (
	"testing"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
)

func TestFuncCheckRequiredServiceFieldsExists(t *testing.T) {
	testCases := []struct {
		testName      string
		service       *v1.Service
		expectedError bool
	}{

		{
			testName: "with annotations and type LoadBalancer",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "nginx",
					Annotations: map[string]string{
						serviceAnnotationPath:     "/foo/bar",
						serviceAnnotationPortName: "http",
					},
				},
				Spec: v1.ServiceSpec{
					Type: "LoadBalancer",
					Ports: []v1.ServicePort{
						{
							Name: "http",
							Port: 80,
						},
					},
				},
			},
			expectedError: false,
		},

		{
			testName: "missing serviceAnnotationPath annotations",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "nginx",
					Annotations: map[string]string{
						serviceAnnotationPortName: "http",
					},
				},
				Spec: v1.ServiceSpec{},
			},
			expectedError: true,
		},

		{
			testName: "missing serviceAnnotationPortName annotations",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "nginx",
					Annotations: map[string]string{
						serviceAnnotationPath: "/foo/bar",
					},
				},
				Spec: v1.ServiceSpec{},
			},
			expectedError: true,
		},

		{
			testName: "type ClusterIP",
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "nginx",
					Annotations: map[string]string{
						serviceAnnotationPath:     "/foo/bar",
						serviceAnnotationPortName: "http",
					},
				},
				Spec: v1.ServiceSpec{
					Type: "ClusterIP",
					Ports: []v1.ServicePort{
						{
							Name: "http",
							Port: 80,
						},
					},
				},
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		err := checkRequiredServiceFieldsExists(tc.service)
		if tc.expectedError == true {
			assert.NotNil(t, err, tc.testName)
		} else {
			assert.Nil(t, err, tc.testName)
		}
	}
}

func TestFuncGetServicePortByName(t *testing.T) {
	testCases := []struct {
		testName        string
		service         *v1.Service
		servicePortName string
		servicePort     *v1.ServicePort
		expectedError   bool
	}{

		{
			testName:        "with matching port",
			servicePortName: "http",
			servicePort: &v1.ServicePort{
				Name: "http",
				Port: 80,
			},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "nginx",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Name: "http",
							Port: 80,
						},
					},
				},
			},
			expectedError: false,
		},

		{
			testName:        "with all annotatitons",
			servicePortName: "http",
			servicePort:     &v1.ServicePort{},
			service: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "nginx",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Name: "foo",
							Port: 80,
						},
					},
				},
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		port := getServicePortByName(tc.servicePortName, tc.service)
		if tc.expectedError == true {
			assert.Nil(t, port, tc.testName)
		} else {
			assert.NotNil(t, port, tc.testName)
			assert.Equal(t, tc.servicePort.Name, port.Name)
			assert.Equal(t, tc.servicePort.Port, port.Port)
		}
	}
}
