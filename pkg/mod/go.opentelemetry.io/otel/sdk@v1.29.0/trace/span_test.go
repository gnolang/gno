// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func TestSetStatus(t *testing.T) {
	tests := []struct {
		name        string
		span        recordingSpan
		code        codes.Code
		description string
		expected    Status
	}{
		{
			"Error and description should overwrite Unset",
			recordingSpan{},
			codes.Error,
			"description",
			Status{Code: codes.Error, Description: "description"},
		},
		{
			"Ok should overwrite Unset and ignore description",
			recordingSpan{},
			codes.Ok,
			"description",
			Status{Code: codes.Ok},
		},
		{
			"Error and description should return error and overwrite description",
			recordingSpan{status: Status{Code: codes.Error, Description: "d1"}},
			codes.Error,
			"d2",
			Status{Code: codes.Error, Description: "d2"},
		},
		{
			"Ok should overwrite error and remove description",
			recordingSpan{status: Status{Code: codes.Error, Description: "d1"}},
			codes.Ok,
			"d2",
			Status{Code: codes.Ok},
		},
		{
			"Error and description should be ignored when already Ok",
			recordingSpan{status: Status{Code: codes.Ok}},
			codes.Error,
			"d2",
			Status{Code: codes.Ok},
		},
		{
			"Ok should be noop when already Ok",
			recordingSpan{status: Status{Code: codes.Ok}},
			codes.Ok,
			"d2",
			Status{Code: codes.Ok},
		},
		{
			"Unset should be noop when already Ok",
			recordingSpan{status: Status{Code: codes.Ok}},
			codes.Unset,
			"d2",
			Status{Code: codes.Ok},
		},
		{
			"Unset should be noop when already Error",
			recordingSpan{status: Status{Code: codes.Error, Description: "d1"}},
			codes.Unset,
			"d2",
			Status{Code: codes.Error, Description: "d1"},
		},
	}

	for i := range tests {
		tc := &tests[i]
		t.Run(tc.name, func(t *testing.T) {
			tc.span.SetStatus(tc.code, tc.description)
			assert.Equal(t, tc.expected, tc.span.status)
		})
	}
}

func TestTruncateAttr(t *testing.T) {
	const key = "key"

	strAttr := attribute.String(key, "value")
	strSliceAttr := attribute.StringSlice(key, []string{"value-0", "value-1"})

	tests := []struct {
		limit      int
		attr, want attribute.KeyValue
	}{
		{
			limit: -1,
			attr:  strAttr,
			want:  strAttr,
		},
		{
			limit: -1,
			attr:  strSliceAttr,
			want:  strSliceAttr,
		},
		{
			limit: 0,
			attr:  attribute.Bool(key, true),
			want:  attribute.Bool(key, true),
		},
		{
			limit: 0,
			attr:  attribute.BoolSlice(key, []bool{true, false}),
			want:  attribute.BoolSlice(key, []bool{true, false}),
		},
		{
			limit: 0,
			attr:  attribute.Int(key, 42),
			want:  attribute.Int(key, 42),
		},
		{
			limit: 0,
			attr:  attribute.IntSlice(key, []int{42, -1}),
			want:  attribute.IntSlice(key, []int{42, -1}),
		},
		{
			limit: 0,
			attr:  attribute.Int64(key, 42),
			want:  attribute.Int64(key, 42),
		},
		{
			limit: 0,
			attr:  attribute.Int64Slice(key, []int64{42, -1}),
			want:  attribute.Int64Slice(key, []int64{42, -1}),
		},
		{
			limit: 0,
			attr:  attribute.Float64(key, 42),
			want:  attribute.Float64(key, 42),
		},
		{
			limit: 0,
			attr:  attribute.Float64Slice(key, []float64{42, -1}),
			want:  attribute.Float64Slice(key, []float64{42, -1}),
		},
		{
			limit: 0,
			attr:  strAttr,
			want:  attribute.String(key, ""),
		},
		{
			limit: 0,
			attr:  strSliceAttr,
			want:  attribute.StringSlice(key, []string{"", ""}),
		},
		{
			limit: 0,
			attr:  attribute.Stringer(key, bytes.NewBufferString("value")),
			want:  attribute.String(key, ""),
		},
		{
			limit: 1,
			attr:  strAttr,
			want:  attribute.String(key, "v"),
		},
		{
			limit: 1,
			attr:  strSliceAttr,
			want:  attribute.StringSlice(key, []string{"v", "v"}),
		},
		{
			limit: 5,
			attr:  strAttr,
			want:  strAttr,
		},
		{
			limit: 7,
			attr:  strSliceAttr,
			want:  strSliceAttr,
		},
		{
			limit: 6,
			attr:  attribute.StringSlice(key, []string{"value", "value-1"}),
			want:  attribute.StringSlice(key, []string{"value", "value-"}),
		},
		{
			limit: 128,
			attr:  strAttr,
			want:  strAttr,
		},
		{
			limit: 128,
			attr:  strSliceAttr,
			want:  strSliceAttr,
		},
		{
			// This tests the ordinary safeTruncate().
			limit: 10,
			attr:  attribute.String(key, "€€€€"), // 3 bytes each
			want:  attribute.String(key, "€€€"),
		},
		{
			// This tests truncation with an invalid UTF-8 input.
			//
			// Note that after removing the invalid rune,
			// the string is over length and still has to
			// be truncated on a code point boundary.
			limit: 10,
			attr:  attribute.String(key, "€"[0:2]+"hello€€"), // corrupted first rune, then over limit
			want:  attribute.String(key, "hello€"),
		},
		{
			// This tests the fallback to invalidTruncate()
			// where after validation the string does not require
			// truncation.
			limit: 6,
			attr:  attribute.String(key, "€"[0:2]+"hello"), // corrupted first rune, then not over limit
			want:  attribute.String(key, "hello"),
		},
	}

	for _, test := range tests {
		name := fmt.Sprintf("%s->%s(limit:%d)", test.attr.Key, test.attr.Value.Emit(), test.limit)
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.want, truncateAttr(test.limit, test.attr))
		})
	}
}

func TestLogDropAttrs(t *testing.T) {
	orig := logDropAttrs
	t.Cleanup(func() { logDropAttrs = orig })

	var called bool
	logDropAttrs = func() { called = true }

	s := &recordingSpan{}
	s.addDroppedAttr(1)
	assert.True(t, called, "logDropAttrs not called")

	called = false
	s.addDroppedAttr(1)
	assert.False(t, called, "logDropAttrs called multiple times for same Span")
}

func BenchmarkRecordingSpanSetAttributes(b *testing.B) {
	var attrs []attribute.KeyValue
	for i := 0; i < 100; i++ {
		attr := attribute.String(fmt.Sprintf("hello.attrib%d", i), fmt.Sprintf("goodbye.attrib%d", i))
		attrs = append(attrs, attr)
	}

	ctx := context.Background()
	for _, limit := range []bool{false, true} {
		b.Run(fmt.Sprintf("WithLimit/%t", limit), func(b *testing.B) {
			b.ReportAllocs()
			sl := NewSpanLimits()
			if limit {
				sl.AttributeCountLimit = 50
			}
			tp := NewTracerProvider(WithSampler(AlwaysSample()), WithSpanLimits(sl))
			tracer := tp.Tracer("tracer")

			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				_, span := tracer.Start(ctx, "span")
				span.SetAttributes(attrs...)
				span.End()
			}
		})
	}
}

func BenchmarkSpanEnd(b *testing.B) {
	tracer := NewTracerProvider().Tracer("")
	ctx := trace.ContextWithSpanContext(context.Background(), trace.SpanContext{})

	spans := make([]trace.Span, b.N)
	for i := 0; i < b.N; i++ {
		_, span := tracer.Start(ctx, "")
		spans[i] = span
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spans[i].End()
	}
}
