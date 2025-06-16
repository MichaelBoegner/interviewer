package chatgpt

import (
	"fmt"
	"os"
	"strings"
)

type ChatGPTResponse struct {
	Topic        string `json:"topic"`
	Subtopic     string `json:"subtopic"`
	Question     string `json:"question"`
	Score        int    `json:"score"`
	Feedback     string `json:"feedback"`
	NextQuestion string `json:"next_question"`
	NextTopic    string `json:"next_topic"`
	NextSubtopic string `json:"next_subtopic"`
}

type OpenAIClient struct {
	APIKey string
}

type OpenAIError struct {
	StatusCode int
	Message    string
}

func (e *OpenAIError) Error() string {
	return fmt.Sprintf("OpenAI error %d: %s", e.StatusCode, e.Message)
}

func NewOpenAI() *OpenAIClient {
	return &OpenAIClient{
		APIKey: os.Getenv("OPENAI_API_KEY"),
	}
}

func BuildPrompt(completedTopics []string, currentTopic string, questionNumber int) string {
	return fmt.Sprintf(`You are conducting a structured, coding-language-agnostic, backend development interview. 
The interview follows **six topics in this order**:

1. **Introduction**
2. **Coding**
3. **System Design**
4. **Databases**
5. **Behavioral**
6. **General Backend Knowledge**

You have already covered the following topics: %s.
You are currently on the topic: %s. 

This is question number %d out of 2 for this topic.

**Rules:**
- Ask **exactly 2 questions per topic** before moving to the next.
- Do **not** skip or reorder topics.
- You only have access to the current topic’s conversation history. Always refer to the current topic and topic list order listed above. 
- Format responses as **valid JSON only** (no explanations or extra text).

**If candidate says 'I don't know':**
- Assign **score: 1** and provide minimal feedback.
- Move to the next question.

**JSON Response Format:**
{
    "topic": "current topic",
    "subtopic": "current subtopic",
    "question": "previous question",
    "score": the score (1-10) you think the previous answer deserves. Treat a score of 7 as the minimum passing threshold. Only give 8–10 for answers that are complete, technically sound, and reflect senior-level expertise. Use scores 1–6 freely to reflect any gaps, vagueness, or missed edge cases. Default to 0 if no score is possible,
    "feedback": "Provide extensive, hyper-critical, detailed feedback. Analyze the answer thoroughly: identify strengths, but scrutinize for any gaps in logic, coverage, or technical depth. If anything is missing, vague, or glossed over, call it out. Hold them to a high bar—clarity, completeness, edge cases, best practices, and tradeoffs. End with one specific improvement they should focus on next time.",
    "next_topic": "Advance to the next topic ONLY if this is the second question. Otherwise, stay on the current topic.",
    "next_subtopic": "next subtopic",
    "next_question": "next question"
}`, strings.Join(completedTopics, ", "), currentTopic, questionNumber)
}

type AIClient interface {
	GetChatGPTResponseInterview(prompt string) (*ChatGPTResponse, error)
	GetChatGPTResponseConversation(conversationHistory []map[string]string) (*ChatGPTResponse, error)
}
