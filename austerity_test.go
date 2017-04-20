package unilog

import (
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseAusterityLevel(t *testing.T) {
	type ParseAusterityTestCase struct {
		Contents      string
		Expected      AusterityLevel
		ExpectedError error
	}

	cases := []ParseAusterityTestCase{
		{
			Contents: "Sheddable",
			Expected: Sheddable,
		},
		{
			Contents: "SheddablePlus",
			Expected: SheddablePlus,
		},
		{
			Contents: "Critical",
			Expected: Critical,
		},
		{
			Contents: "CriticalPlus",
			Expected: CriticalPlus,
		},
		{
			Contents:      "InvalidAusterity",
			Expected:      Sheddable,
			ExpectedError: InvalidAusterityLevel,
		},
		{
			// test case-insensitivity
			Contents: "sHedDaBlePlUs",
			Expected: SheddablePlus,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Contents, func(t *testing.T) {
			l, err := ParseLevel(strings.NewReader(tc.Contents))
			if tc.ExpectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, err, InvalidAusterityLevel)
			}
			assert.Equal(t, tc.Expected, l)
		})
	}
}

// Test that SendAusterityLevel still works properly
// if loading the file returns an error
func TestLoadLevelError(t *testing.T) {
	// For tests, we don't actually want to implement a delay
	// Since both channels will always be ready to send,
	// the scheduler will choose pseudorandomly between the two
	CacheInterval = 0 * time.Millisecond

	// Assert that the file doesn't actually exist, so it should return an error
	// of type *os.PathError
	l, err := LoadLevel()

	// Check the default austerity level
	// For now, if we encountered an error, we should default
	// to the lowest possible level.
	// This may change at some point later.
	assert.Equal(t, Sheddable, l)
	assert.Error(t, err)
	assert.IsType(t, &os.PathError{}, err)

	go SendSystemAusterityLevel()

	for i := 0; i < 10; i++ {
		lvl := <-SystemAusterityLevel

		assert.Equal(t, Sheddable, lvl)
	}
}

func TestCriticality(t *testing.T) {
	type CriticalityTestCase struct {
		name  string
		line  string
		level AusterityLevel
	}
	cases := []CriticalityTestCase{
		{
			name: "CANONICAL-API-LINE",
			// actual CANONICAL-API-LINE, with explicit merchant-identifying tokens removed
			line:  `[2016-11-10 19:18:05.844100] [98381|f1.northwest-1.apiori.com/EzBDuA4iNq-2631925524 85137cc252d87354>e9b8c49860f01f15] CANONICAL-API-LINE: api_method=AccountRetrieveMethod content_type="application/x-www-form-urlencoded" created=1478805073.5253563 http_method=GET ip="54.210.12.235" path="/v1/accounts/acct_xxxxxxxxxxxxxxxx" user_agent="Stripe/v1 RubyBindings/1.31.0" request_id=req_xxxxxxxxxxxxxx response_stripe_version="2016-03-07" status=200 merchant=acct_xxxxxxxxxxxxx merchant__created=1474909410.2272582 merchant__stripe_version="2016-07-06" logical_shard=maindb_0030 physical_shard=shard_k key_id=mk_xxxxxxxxxxxxx livemode=true perms=admin perms_used="limited_account_read,account_admin_read" application=ca_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx page_mode=none idempotency_work_done=true limiting_considered=true describe_duration=0.175 duration=0.234 end_to_end_duration=0.236 lock_duration=0.002 wabbit_duration=0.001 write_event_duration=0.002 publish_kafka_duration=0.009 cpu_duration=0.066 bindings_language=ruby bindings_language_version="2.2.3 p173 (2015-08-18)" bindings_version="1.31.0" git_revision=5e00cf28 tls_cipher="ECDHE-RSA-AES128-GCM-SHA256" tls_version="TLSv1.2" alloc_count=54187 gc_duration=0.005 duration_publish=0.001`,
			level: CriticalPlus,
		},
		{
			name: "CANONICAL-API-LINE",
			// actual CANONICAL-API-LINE, with explicit merchant-identifying tokens removed
			// this should be criticalplus, despite clevel=sheddable being set
			line:  `[2016-11-10 19:18:05.844100] [98381|f1.northwest-1.apiori.com/EzBDuA4iNq-2631925524 85137cc252d87354>e9b8c49860f01f15] CANONICAL-API-LINE: api_method=AccountRetrieveMethod content_type="application/x-www-form-urlencoded" created=1478805073.5253563 http_method=GET ip="54.210.12.235" path="/v1/accounts/acct_xxxxxxxxxxxxxxxx" user_agent="Stripe/v1 RubyBindings/1.31.0" request_id=req_xxxxxxxxxxxxxx response_stripe_version="2016-03-07" status=200 merchant=acct_xxxxxxxxxxxxx merchant__created=1474909410.2272582 merchant__stripe_version="2016-07-06" logical_shard=maindb_0030 physical_shard=shard_k key_id=mk_xxxxxxxxxxxxx livemode=true perms=admin perms_used="limited_account_read,account_admin_read" application=ca_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx page_mode=none idempotency_work_done=true limiting_considered=true describe_duration=0.175 duration=0.234 end_to_end_duration=0.236 lock_duration=0.002 wabbit_duration=0.001 write_event_duration=0.002 publish_kafka_duration=0.009 cpu_duration=0.066 bindings_language=ruby bindings_language_version="2.2.3 p173 (2015-08-18)" bindings_version="1.31.0" git_revision=5e00cf28 tls_cipher="ECDHE-RSA-AES128-GCM-SHA256" tls_version="TLSv1.2" alloc_count=54187 gc_duration=0.005 duration_publish=0.001 clevel=sheddable`,
			level: CriticalPlus,
		},
		{
			name: "CANONICAL-ADMIN-LINE",
			// actual CANONICAL-ADMIN-LINE, with user-identifying tokens removed
			line:  `[2016-11-10 19:10:49.230930] [22560|adminbox--04ec81f3361370d7f.northwest.stripe.io/kUku-rmgfZ-349 0000000000000000>a020f53ed1dd83ef] CANONICAL-ADMIN-LINE: path="/fonts/glyphicons-halflings-regular.woff" http_method=GET referer="/css/bootstrap3.min.css" response_content_type="application/octet-stream" status=200 duration=0.003100368 stripe_user_full_name="Foo Bar" stripe_user=foo stripe_user_team="user-operations"`,
			level: CriticalPlus,
		},
		{
			name:  "PlainOldLogLine",
			line:  `[2016-11-10 19:01:02.461489] [21515|adminbox--04ec81f3361370d7f.northwest.stripe.io/kUku-wvrZK-28 0000000000000000>93e612b5bd9b69eb] HTTP response headers: Content-Type="text/html;charset=utf-8" Content-Security-Policy="default-src 'self' 'unsafe-eval' https://*.stripe.com https://*.stripe.io:* https://*.stripe.me https://*.google.com https://*.googleapis.com https://*.gstatic.com http://*.mapbox.com;font-src 'self' https://*.stripe.com https://*.stripe.io:* https://*.stripe.me https://*.gstatic.com data:;style-src 'self' 'unsafe-inline' https://*.stripe.com https://*.google.com https://*.googleapis.com https://*.typography.com https://*.mapbox.com;img-src 'self' https://stripe-underwriting-documents.s3.amazonaws.com https://stripe-passport-uploads.s3.amazonaws.com https://*.stripe.com https://*.stripe.io:* https://*.stripe.me https://stripe.com https://*.google.com https://*.googleapis.com https://*.gstatic.com http://*.mapbox.com https://*.mapbox.com data: https://*.stripe.com https://*.stripe.io:* https://*.stripe.me blob:;report-uri https://admin.corp.stripe.com/security/csp-report;frame-src 'self' https://adminsearch.corp.stripe.com https://nagios.northwest.corp.stripe.com https://nagios.east.corp.stripe.com https://tiller.corp.stripe.com https://viz.corp.stripe.com;" X-Content-Security-Policy="default-src 'self' 'unsafe-eval' https://*.stripe.com https://*.stripe.io:* https://*.stripe.me https://*.google.com https://*.googleapis.com https://*.gstatic.com http://*.mapbox.com;font-src 'self' https://*.stripe.com https://*.stripe.io:* https://*.stripe.me https://*.gstatic.com data:;style-src 'self' 'unsafe-inline' https://*.stripe.com https://*.google.com https://*.googleapis.com https://*.typography.com https://*.mapbox.com;img-src 'self' https://stripe-underwriting-documents.s3.amazonaws.com https://stripe-passport-uploads.s3.amazonaws.com https://*.stripe.com https://*.stripe.io:* https://*.stripe.me https://stripe.com https://*.google.com https://*.googleapis.com https://*.gstatic.com http://*.mapbox.com https://*.mapbox.com data: https://*.stripe.com https://*.stripe.io:* https://*.stripe.me blob:;report-uri https://admin.corp.stripe.com/security/csp-report;frame-src 'self' https://adminsearch.corp.stripe.com https://nagios.northwest.corp.stripe.com https://nagios.east.corp.stripe.com https://tiller.corp.stripe.com https://viz.corp.stripe.com;" Content-Length="10879" Set`,
			level: SheddablePlus,
		},
		{
			name:  "PlainOldLogLineWithClevelUppercase",
			line:  `[2016-11-10 19:01:02.461489] [21515|adminbox--04ec81f3361370d7f.northwest.stripe.io/kUku-wvrZK-28 0000000000000000>93e612b5bd9b69eb] HTTP response headers: Content-Type="text/html;charset=utf-8" Content-Security-Policy="default-src 'self' 'unsafe-eval' https://*.stripe.com https://*.stripe.io:* https://*.stripe.me https://*.google.com https://*.googleapis.com https://*.gstatic.com http://*.mapbox.com;font-src 'self' https://*.stripe.com https://*.stripe.io:* https://*.stripe.me https://*.gstatic.com data:;style-src 'self' 'unsafe-inline' https://*.stripe.com https://*.google.com https://*.googleapis.com https://*.typography.com https://*.mapbox.com;img-src 'self' https://stripe-underwriting-documents.s3.amazonaws.com https://stripe-passport-uploads.s3.amazonaws.com https://*.stripe.com https://*.stripe.io:* https://*.stripe.me https://stripe.com https://*.google.com https://*.googleapis.com https://*.gstatic.com http://*.mapbox.com https://*.mapbox.com data: https://*.stripe.com https://*.stripe.io:* https://*.stripe.me blob:;report-uri https://admin.corp.stripe.com/security/csp-report;frame-src 'self' https://adminsearch.corp.stripe.com https://nagios.northwest.corp.stripe.com https://nagios.east.corp.stripe.com https://tiller.corp.stripe.com https://viz.corp.stripe.com;" X-Content-Security-Policy="default-src 'self' 'unsafe-eval' https://*.stripe.com https://*.stripe.io:* https://*.stripe.me https://*.google.com https://*.googleapis.com https://*.gstatic.com http://*.mapbox.com;font-src 'self' https://*.stripe.com https://*.stripe.io:* https://*.stripe.me https://*.gstatic.com data:;style-src 'self' 'unsafe-inline' https://*.stripe.com https://*.google.com https://*.googleapis.com https://*.typography.com https://*.mapbox.com;img-src 'self' https://stripe-underwriting-documents.s3.amazonaws.com https://stripe-passport-uploads.s3.amazonaws.com https://*.stripe.com https://*.stripe.io:* https://*.stripe.me https://stripe.com https://*.google.com https://*.googleapis.com https://*.gstatic.com http://*.mapbox.com https://*.mapbox.com data: https://*.stripe.com https://*.stripe.io:* https://*.stripe.me blob:;report-uri https://admin.corp.stripe.com/security/csp-report;frame-src 'self' https://adminsearch.corp.stripe.com https://nagios.northwest.corp.stripe.com https://nagios.east.corp.stripe.com https://tiller.corp.stripe.com https://viz.corp.stripe.com;" Content-Length="10879" Set [clevel: Critical]`,
			level: Critical,
		},
		{
			name:  "PlainOldLogLineWithClevelLowercase",
			line:  `[2016-11-10 19:01:02.461489] [21515|adminbox--04ec81f3361370d7f.northwest.stripe.io/kUku-wvrZK-28 0000000000000000>93e612b5bd9b69eb] HTTP response headers: Content-Type="text/html;charset=utf-8" Content-Security-Policy="default-src 'self' 'unsafe-eval' https://*.stripe.com https://*.stripe.io:* https://*.stripe.me https://*.google.com https://*.googleapis.com https://*.gstatic.com http://*.mapbox.com;font-src 'self' https://*.stripe.com https://*.stripe.io:* https://*.stripe.me https://*.gstatic.com data:;style-src 'self' 'unsafe-inline' https://*.stripe.com https://*.google.com https://*.googleapis.com https://*.typography.com https://*.mapbox.com;img-src 'self' https://stripe-underwriting-documents.s3.amazonaws.com https://stripe-passport-uploads.s3.amazonaws.com https://*.stripe.com https://*.stripe.io:* https://*.stripe.me https://stripe.com https://*.google.com https://*.googleapis.com https://*.gstatic.com http://*.mapbox.com https://*.mapbox.com data: https://*.stripe.com https://*.stripe.io:* https://*.stripe.me blob:;report-uri https://admin.corp.stripe.com/security/csp-report;frame-src 'self' https://adminsearch.corp.stripe.com https://nagios.northwest.corp.stripe.com https://nagios.east.corp.stripe.com https://tiller.corp.stripe.com https://viz.corp.stripe.com;" X-Content-Security-Policy="default-src 'self' 'unsafe-eval' https://*.stripe.com https://*.stripe.io:* https://*.stripe.me https://*.google.com https://*.googleapis.com https://*.gstatic.com http://*.mapbox.com;font-src 'self' https://*.stripe.com https://*.stripe.io:* https://*.stripe.me https://*.gstatic.com data:;style-src 'self' 'unsafe-inline' https://*.stripe.com https://*.google.com https://*.googleapis.com https://*.typography.com https://*.mapbox.com;img-src 'self' https://stripe-underwriting-documents.s3.amazonaws.com https://stripe-passport-uploads.s3.amazonaws.com https://*.stripe.com https://*.stripe.io:* https://*.stripe.me https://stripe.com https://*.google.com https://*.googleapis.com https://*.gstatic.com http://*.mapbox.com https://*.mapbox.com data: https://*.stripe.com https://*.stripe.io:* https://*.stripe.me blob:;report-uri https://admin.corp.stripe.com/security/csp-report;frame-src 'self' https://adminsearch.corp.stripe.com https://nagios.northwest.corp.stripe.com https://nagios.east.corp.stripe.com https://tiller.corp.stripe.com https://viz.corp.stripe.com;" Content-Length="10879" Set [clevel: critical]`,
			level: Critical,
		},
		{
			name:  "LogLineWithClevelChalk",
			line:  `[2016-11-10 20:02:01.932272] [24607|adminbox--04ec81f3361370d7f.northwest.stripe.io/kUku-WdiJA-3204 0000000000000000>831e61790017a475] Showed dispute tier for merchant: merchant=acct_1iQSUkTgZeIC7dPrB0qj tier=tier0 clevel=criticalplus`,
			level: CriticalPlus,
		},
	}

	for i, tc := range cases {
		name := strconv.Itoa(i)
		if tc.name != "" {
			name = tc.name
		}
		t.Run(name, func(t *testing.T) {
			l := criticality(tc.line)
			assert.Equal(t, tc.level.String(), l.String())
		})
	}
}
