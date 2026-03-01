// Copyright 2025 The Serverless Workflow Specification Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentDefinition_Defaults(t *testing.T) {
	agent := &AgentDefinition{
		Role: "Analyst",
		Goal: "Analyze data",
	}

	assert.Equal(t, 10, agent.GetMaxIterations())
	assert.Equal(t, 0.7, agent.GetTemperature())
	assert.False(t, agent.IsRef())
}

func TestAgentDefinition_Ref(t *testing.T) {
	agent := &AgentDefinition{
		Ref: "senior-analyst",
	}
	assert.True(t, agent.IsRef())
	assert.Equal(t, "senior-analyst", agent.Ref)
}

func TestAgentDefinition_Merge(t *testing.T) {
	baseTemp := 0.5
	base := &AgentDefinition{
		Role:          "Data Analyst",
		Goal:          "Analyze data",
		Backstory:     "Expert analyst",
		LLM:           "openai/gpt-4o",
		Temperature:   &baseTemp,
		Tools:         []string{"hubspot:search"},
		MaxIterations: 10,
		Memory:        false,
	}

	overrideTemp := 0.3
	override := &AgentDefinition{
		Ref:         "data-analyst", // should be cleared in merged result
		LLM:         "anthropic/claude-sonnet-4",
		Temperature: &overrideTemp,
		Tools:       []string{"slack:send"},
	}

	merged := override.Merge(base)

	// Base values preserved
	assert.Equal(t, "Data Analyst", merged.Role)
	assert.Equal(t, "Analyze data", merged.Goal)
	assert.Equal(t, "Expert analyst", merged.Backstory)
	assert.Equal(t, 10, merged.MaxIterations)

	// Overridden values
	assert.Equal(t, "anthropic/claude-sonnet-4", merged.LLM)
	assert.Equal(t, 0.3, *merged.Temperature)

	// Tools are appended
	assert.Equal(t, []string{"hubspot:search", "slack:send"}, merged.Tools)

	// Ref is cleared on merged result
	assert.Empty(t, merged.Ref)
	assert.False(t, merged.IsRef())

	// Original base is not modified
	assert.Equal(t, "openai/gpt-4o", base.LLM)
}

func TestAgentDefinition_Merge_EmptyOverride(t *testing.T) {
	base := &AgentDefinition{
		Role: "Writer",
		Goal: "Write reports",
		LLM:  "openai/gpt-4o",
	}

	override := &AgentDefinition{Ref: "writer"}
	merged := override.Merge(base)

	assert.Equal(t, "Writer", merged.Role)
	assert.Equal(t, "Write reports", merged.Goal)
	assert.Equal(t, "openai/gpt-4o", merged.LLM)
}

func TestAgentDefinition_CustomValues(t *testing.T) {
	temp := 0.3
	agent := &AgentDefinition{
		Role:          "Writer",
		Goal:          "Write reports",
		MaxIterations: 25,
		Temperature:   &temp,
	}

	assert.Equal(t, 25, agent.GetMaxIterations())
	assert.Equal(t, 0.3, agent.GetTemperature())
}

func TestAgentDefinition_JSON_RoundTrip(t *testing.T) {
	temp := 0.5
	agent := &AgentDefinition{
		Role:            "Senior Researcher",
		Goal:            "Find information",
		Backstory:       "Expert researcher",
		LLM:             "openai/gpt-4o",
		Temperature:     &temp,
		MaxTokens:       intPtr(4096),
		Tools:           []string{"web:search", "web:scrape"},
		AllowDelegation: true,
		MaxIterations:   15,
		Memory:          true,
		MemoryScope:     "research",
		ResponseFormat:  "json",
	}

	data, err := json.Marshal(agent)
	require.NoError(t, err)

	var decoded AgentDefinition
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, agent.Role, decoded.Role)
	assert.Equal(t, agent.Goal, decoded.Goal)
	assert.Equal(t, agent.Backstory, decoded.Backstory)
	assert.Equal(t, agent.LLM, decoded.LLM)
	assert.Equal(t, *agent.Temperature, *decoded.Temperature)
	assert.Equal(t, *agent.MaxTokens, *decoded.MaxTokens)
	assert.Equal(t, agent.Tools, decoded.Tools)
	assert.Equal(t, agent.AllowDelegation, decoded.AllowDelegation)
	assert.Equal(t, agent.MaxIterations, decoded.MaxIterations)
	assert.Equal(t, agent.Memory, decoded.Memory)
	assert.Equal(t, agent.MemoryScope, decoded.MemoryScope)
	assert.Equal(t, agent.ResponseFormat, decoded.ResponseFormat)
}

func TestAgentTask_JSON_RoundTrip(t *testing.T) {
	task := &AgentTask{
		Agent:               "analyst",
		TaskDescription:     "Analyze the customer data",
		ExpectedOutput:      "A structured analysis report",
		Context:             []string{"fetchData"},
		Tools:               []string{"extra:tool"},
		Guardrail:           "${ $count(output.items) > 0 }",
		GuardrailMaxRetries: 5,
	}

	data, err := json.Marshal(task)
	require.NoError(t, err)

	var decoded AgentTask
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "analyst", decoded.Agent)
	assert.Equal(t, "Analyze the customer data", decoded.TaskDescription)
	assert.Equal(t, "A structured analysis report", decoded.ExpectedOutput)
	assert.Equal(t, []string{"fetchData"}, decoded.Context)
	assert.Equal(t, []string{"extra:tool"}, decoded.Tools)
	assert.Equal(t, "${ $count(output.items) > 0 }", decoded.Guardrail)
	assert.Equal(t, 5, decoded.GuardrailMaxRetries)
}

func TestAgentTask_GetGuardrailMaxRetries_Default(t *testing.T) {
	task := &AgentTask{Agent: "test", TaskDescription: "test"}
	assert.Equal(t, 3, task.GetGuardrailMaxRetries())
}

func TestAgentTask_GetBase(t *testing.T) {
	task := &AgentTask{
		TaskBase: TaskBase{
			Then: &FlowDirective{Value: "end"},
		},
		Agent:           "test",
		TaskDescription: "test task",
	}

	base := task.GetBase()
	require.NotNil(t, base)
	assert.Equal(t, "end", base.Then.Value)
}

func TestAgentTask_Registered(t *testing.T) {
	constructor, exists := GetTaskConstructor("agent")
	assert.True(t, exists, "agent task type should be registered")

	task := constructor()
	_, ok := task.(*AgentTask)
	assert.True(t, ok, "constructor should return *AgentTask")
}

func TestAgentTask_UnmarshalFromTaskList(t *testing.T) {
	jsonData := `[
		{
			"analyzeData": {
				"agent": "analyst",
				"task": "Analyze the contacts",
				"expected_output": "A report",
				"context": ["fetchData"],
				"then": "end"
			}
		}
	]`

	var taskList TaskList
	err := json.Unmarshal([]byte(jsonData), &taskList)
	require.NoError(t, err)
	require.Len(t, taskList, 1)

	assert.Equal(t, "analyzeData", taskList[0].Key)

	agentTask := taskList[0].AsAgentTask()
	require.NotNil(t, agentTask, "should cast to AgentTask")
	assert.Equal(t, "analyst", agentTask.Agent)
	assert.Equal(t, "Analyze the contacts", agentTask.TaskDescription)
	assert.Equal(t, "A report", agentTask.ExpectedOutput)
	assert.Equal(t, []string{"fetchData"}, agentTask.Context)
	assert.Equal(t, "end", agentTask.GetBase().Then.Value)
}

func TestUse_Agents_JSON(t *testing.T) {
	jsonData := `{
		"agents": {
			"analyst": {
				"role": "Data Analyst",
				"goal": "Analyze data",
				"llm": "openai/gpt-4o",
				"tools": ["hubspot:search-contacts"]
			},
			"writer": {
				"role": "Content Writer",
				"goal": "Write reports",
				"llm": "anthropic/claude-sonnet-4"
			}
		}
	}`

	var use Use
	err := json.Unmarshal([]byte(jsonData), &use)
	require.NoError(t, err)
	require.Len(t, use.Agents, 2)

	analyst := use.Agents["analyst"]
	require.NotNil(t, analyst)
	assert.Equal(t, "Data Analyst", analyst.Role)
	assert.Equal(t, "Analyze data", analyst.Goal)
	assert.Equal(t, "openai/gpt-4o", analyst.LLM)
	assert.Equal(t, []string{"hubspot:search-contacts"}, analyst.Tools)

	writer := use.Agents["writer"]
	require.NotNil(t, writer)
	assert.Equal(t, "Content Writer", writer.Role)
}

func TestAsAgentTask_NilTaskItem(t *testing.T) {
	var ti *TaskItem
	assert.Nil(t, ti.AsAgentTask())
}

func TestAsAgentTask_WrongType(t *testing.T) {
	ti := &TaskItem{
		Key:  "test",
		Task: &SetTask{},
	}
	assert.Nil(t, ti.AsAgentTask())
}

func TestUse_Agents_Ref_JSON(t *testing.T) {
	jsonData := `{
		"agents": {
			"analyst": {
				"ref": "senior-data-analyst",
				"llm": "anthropic/claude-sonnet-4",
				"temperature": 0.2
			}
		}
	}`

	var use Use
	err := json.Unmarshal([]byte(jsonData), &use)
	require.NoError(t, err)
	require.Len(t, use.Agents, 1)

	analyst := use.Agents["analyst"]
	require.NotNil(t, analyst)
	assert.True(t, analyst.IsRef())
	assert.Equal(t, "senior-data-analyst", analyst.Ref)
	assert.Equal(t, "anthropic/claude-sonnet-4", analyst.LLM)
	assert.Equal(t, 0.2, *analyst.Temperature)
	assert.Empty(t, analyst.Role) // not required when ref is set
}

func intPtr(i int) *int {
	return &i
}
