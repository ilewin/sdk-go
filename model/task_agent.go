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

// AgentDefinition defines an AI agent's configuration.
// Supports two modes:
//   - Inline: all fields set directly in YAML (role+goal required)
//   - Reference: set Ref to a stored agent name, other fields act as overrides
type AgentDefinition struct {
	// Ref references a stored agent definition by name (from agent_definitions table).
	// When set, other fields act as overrides on top of the stored definition.
	Ref string `json:"ref,omitempty"`

	// Role is the agent's functional role (e.g., "Senior Data Analyst").
	Role string `json:"role,omitempty"`
	// Goal is the objective the agent aims to achieve.
	Goal string `json:"goal,omitempty"`
	// Backstory provides background context for the agent's persona.
	Backstory string `json:"backstory,omitempty"`

	// LLM specifies the model in "provider/model" format (e.g., "openai/gpt-4o").
	LLM string `json:"llm,omitempty"`
	// Temperature controls LLM output randomness (0.0-2.0).
	Temperature *float64 `json:"temperature,omitempty"`
	// MaxTokens limits the LLM output length.
	MaxTokens *int `json:"max_tokens,omitempty"`
	// Credential references a secret in use.secrets for the LLM API key.
	// If empty, the tenant's default credential for the provider is used.
	Credential string `json:"credential,omitempty"`

	// Tools lists callable tool references (e.g., "hubspot:search-contacts@conn-1").
	Tools []string `json:"tools,omitempty"`
	// AllowDelegation enables delegate_work and ask_question built-in tools.
	AllowDelegation bool `json:"allow_delegation,omitempty"`

	// MaxIterations limits the reasoning loop iterations (default: 10).
	MaxIterations int `json:"max_iterations,omitempty"`
	// MaxExecutionTime is a hard timeout for agent execution.
	MaxExecutionTime *Duration `json:"max_execution_time,omitempty"`

	// Memory enables memory recall/storage across executions.
	Memory bool `json:"memory,omitempty"`
	// MemoryScope controls memory isolation scope (default: agent role).
	MemoryScope string `json:"memory_scope,omitempty"`

	// SystemPrompt overrides the default system prompt template.
	// When set, it takes precedence over role/goal/backstory for prompt generation.
	SystemPrompt string `json:"system_prompt,omitempty"`
	// ResponseFormat controls output format: "text" (default) or "json".
	ResponseFormat string `json:"response_format,omitempty" validate:"omitempty,oneof=text json"`
}

// IsRef returns true if this definition references a stored agent.
func (a *AgentDefinition) IsRef() bool {
	return a.Ref != ""
}

// Merge applies non-zero override fields from the receiver on top of a base definition.
// Used to merge inline overrides on top of a DB-stored agent.
// Returns a new AgentDefinition with the merged result.
func (a *AgentDefinition) Merge(base *AgentDefinition) *AgentDefinition {
	merged := *base // shallow copy of base
	merged.Ref = "" // clear ref on the resolved copy

	if a.Role != "" {
		merged.Role = a.Role
	}
	if a.Goal != "" {
		merged.Goal = a.Goal
	}
	if a.Backstory != "" {
		merged.Backstory = a.Backstory
	}
	if a.LLM != "" {
		merged.LLM = a.LLM
	}
	if a.Temperature != nil {
		merged.Temperature = a.Temperature
	}
	if a.MaxTokens != nil {
		merged.MaxTokens = a.MaxTokens
	}
	if a.Credential != "" {
		merged.Credential = a.Credential
	}
	if len(a.Tools) > 0 {
		merged.Tools = append(merged.Tools, a.Tools...)
	}
	if a.AllowDelegation {
		merged.AllowDelegation = true
	}
	if a.MaxIterations > 0 {
		merged.MaxIterations = a.MaxIterations
	}
	if a.MaxExecutionTime != nil {
		merged.MaxExecutionTime = a.MaxExecutionTime
	}
	if a.Memory {
		merged.Memory = true
	}
	if a.MemoryScope != "" {
		merged.MemoryScope = a.MemoryScope
	}
	if a.SystemPrompt != "" {
		merged.SystemPrompt = a.SystemPrompt
	}
	if a.ResponseFormat != "" {
		merged.ResponseFormat = a.ResponseFormat
	}
	return &merged
}

// GetMaxIterations returns MaxIterations or the default of 10.
func (a *AgentDefinition) GetMaxIterations() int {
	if a.MaxIterations > 0 {
		return a.MaxIterations
	}
	return 10
}

// GetTemperature returns the temperature or a default of 0.7.
func (a *AgentDefinition) GetTemperature() float64 {
	if a.Temperature != nil {
		return *a.Temperature
	}
	return 0.7
}

// AgentTask is a DSL task that assigns work to an AI agent.
// The agent field references a definition in workflow.use.agents.
type AgentTask struct {
	TaskBase `json:",inline"`
	// Agent references an agent defined in use.agents by name.
	Agent string `json:"agent" validate:"required"`
	// TaskDescription is the work to be performed (supports expressions).
	TaskDescription string `json:"task" validate:"required"`
	// ExpectedOutput describes what the agent should produce.
	ExpectedOutput string `json:"expected_output,omitempty"`
	// Context lists task keys whose outputs are provided as context.
	Context []string `json:"context,omitempty"`
	// Tools lists additional tool references for this specific task.
	// These supplement the agent's own tools.
	Tools []string `json:"tools,omitempty"`
	// OutputSchema is a JSON Schema for validating structured output.
	OutputSchema map[string]interface{} `json:"output_schema,omitempty"`
	// Guardrail is an expression evaluated against the output.
	// If it evaluates to false, the agent retries.
	Guardrail string `json:"guardrail,omitempty"`
	// GuardrailMaxRetries limits guardrail retry attempts (default: 3).
	GuardrailMaxRetries int `json:"guardrail_max_retries,omitempty"`
}

// GetBase returns the embedded TaskBase.
func (a *AgentTask) GetBase() *TaskBase {
	return &a.TaskBase
}

// GetGuardrailMaxRetries returns GuardrailMaxRetries or the default of 3.
func (a *AgentTask) GetGuardrailMaxRetries() int {
	if a.GuardrailMaxRetries > 0 {
		return a.GuardrailMaxRetries
	}
	return 3
}
