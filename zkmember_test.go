package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHaveEndpoints(t *testing.T) {
	testCases := []struct {
		testName string
		member   *zkMember
		expected bool
	}{

		{
			testName: "",
			expected: true,
			member: &zkMember{
				name: "foo",
				ServiceEndpoint: zkMemberUnite{
					Host: "localhost",
					Port: 123,
				},
				AdditionalEndpoints: Endpoints{
					"http": zkMemberUnite{
						Host: "localhost",
						Port: 123,
					},
				},
			},
		},

		{
			testName: "",
			expected: true,
			member: &zkMember{
				name: "foo",
				AdditionalEndpoints: Endpoints{
					"http": zkMemberUnite{
						Host: "localhost",
						Port: 123,
					},
				},
			},
		},

		{
			testName: "",
			expected: false,
			member: &zkMember{
				name:                "foo",
				AdditionalEndpoints: make(Endpoints),
			},
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, tc.member.anyEndpoints(), tc.testName)
	}
}