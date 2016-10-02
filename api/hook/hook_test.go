package hook

import (
	"testing"

	"github.com/ovh/tat"
	"github.com/stretchr/testify/assert"
)

func TestMatchCriteria(t *testing.T) {

	assert.Equal(t, false, matchCriteria(
		tat.Message{Labels: []tat.Label{{Text: "labelA", Color: "#eeeeee"}}},
		tat.MessageCriteria{Label: "labelB"}),
		"this message should not match")

	assert.Equal(t, true, matchCriteria(
		tat.Message{Labels: []tat.Label{{Text: "labelA", Color: "#eeeeee"}}},
		tat.MessageCriteria{Label: "labelA"}),
		"this message should match")

	assert.Equal(t, false, matchCriteria(
		tat.Message{Labels: []tat.Label{{Text: "labelA", Color: "#eeeeee"}}},
		tat.MessageCriteria{NotLabel: "labelA"}),
		"this message should not match")

	assert.Equal(t, true, matchCriteria(
		tat.Message{Labels: []tat.Label{{Text: "labelA", Color: "#eeeeee"}, {Text: "labelB", Color: "#eeeeee"}}},
		tat.MessageCriteria{AndLabel: "labelA"}),
		"this message should match")

	assert.Equal(t, true, matchCriteria(
		tat.Message{Labels: []tat.Label{{Text: "labelA", Color: "#eeeeee"}, {Text: "labelB", Color: "#eeeeee"}}},
		tat.MessageCriteria{AndLabel: "labelA,labelB"}),
		"this message should match")

	assert.Equal(t, false, matchCriteria(
		tat.Message{Labels: []tat.Label{{Text: "labelA", Color: "#eeeeee"}, {Text: "labelB", Color: "#eeeeee"}}},
		tat.MessageCriteria{AndLabel: "labelA,labelB,labelC"}),
		"this message should not match")

	assert.Equal(t, false, matchCriteria(
		tat.Message{Tags: []string{"tagA"}},
		tat.MessageCriteria{Tag: "tagB"}),
		"this message should not match")

	assert.Equal(t, true, matchCriteria(
		tat.Message{Tags: []string{"tagA"}},
		tat.MessageCriteria{Tag: "tagA"}),
		"this message should match")

	assert.Equal(t, false, matchCriteria(
		tat.Message{Tags: []string{"tagA"}},
		tat.MessageCriteria{NotTag: "tagA"}),
		"this message should not match")

	assert.Equal(t, true, matchCriteria(
		tat.Message{Tags: []string{"tagA", "tagB"}},
		tat.MessageCriteria{AndTag: "tagA"}),
		"this message should match")

	assert.Equal(t, true, matchCriteria(
		tat.Message{Tags: []string{"tagA", "tagB"}},
		tat.MessageCriteria{AndTag: "tagA,tagB"}),
		"this message should match")

	assert.Equal(t, false, matchCriteria(
		tat.Message{Tags: []string{"tagA", "tagB"}},
		tat.MessageCriteria{AndTag: "tagA,tagB,tagC"}),
		"this message should not match")

	assert.Equal(t, true, matchCriteria(
		tat.Message{InReplyOfID: ""},
		tat.MessageCriteria{OnlyMsgRoot: tat.True}),
		"this message should match")

	assert.Equal(t, false, matchCriteria(
		tat.Message{InReplyOfID: "fff"},
		tat.MessageCriteria{OnlyMsgRoot: tat.True}),
		"this message should not match")

	assert.Equal(t, true, matchCriteria(
		tat.Message{Author: tat.Author{Username: "foo"}},
		tat.MessageCriteria{Username: "foo"}),
		"this message should match")

	assert.Equal(t, false, matchCriteria(
		tat.Message{Author: tat.Author{Username: "foo"}},
		tat.MessageCriteria{Username: "bar"}),
		"this message should not match")

	assert.Equal(t, true, matchCriteria(
		tat.Message{Labels: []tat.Label{{Text: "labelA", Color: "#eeeeee"}, {Text: "labelB", Color: "#eeeeee"}}, Tags: []string{"tagA", "tagB", "tagC"}},
		tat.MessageCriteria{AndTag: "tagA,tagB", AndLabel: "labelA,labelB"}),
		"this message should match")
}
