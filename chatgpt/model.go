package chatgpt

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type ChatGPTResponse struct {
	Topic            string   `json:"topic"`
	Subtopic         string   `json:"subtopic"`
	Question         string   `json:"question"`
	Score            int      `json:"score"`
	Feedback         string   `json:"feedback"`
	NextQuestion     string   `json:"next_question"`
	NextTopic        string   `json:"next_topic"`
	NextSubtopic     string   `json:"next_subtopic"`
	Domain           string   `json:"domain"`
	Responsibilities []string `json:"responsibilities"`
	Qualifications   []string `json:"qualifications"`
	TechStack        []string `json:"tech_stack"`
	Level            string   `json:"level"`
}

type OpenAIClient struct {
	APIKey string
	Logger *slog.Logger
}

type OpenAIError struct {
	StatusCode int
	Message    string
}

type JDParsedOutput struct {
	Domain           string   `json:"domain"`
	Responsibilities []string `json:"responsibilities"`
	Qualifications   []string `json:"qualifications"`
	TechStack        []string `json:"tech_stack"`
	Level            string   `json:"level"`
}

func (e *OpenAIError) Error() string {
	return fmt.Sprintf("OpenAI error %d: %s", e.StatusCode, e.Message)
}

func NewOpenAI(logger *slog.Logger) *OpenAIClient {
	return &OpenAIClient{
		APIKey: os.Getenv("OPENAI_API_KEY"),
		Logger: logger,
	}
}

func BuildPrompt(completedTopics []string, currentTopic string, questionNumber int, jdSummary string) string {
	return fmt.Sprintf(`You are conducting a structured, coding-language-agnostic, backend development interview.

**Rules:**
- Ask **exactly 2 questions per topic** before moving to the next.
- Do **not** skip or reorder topics.
- You only have access to the current topic’s conversation history. Always refer to the current topic, topic list order, and question number below.
- **Every question must relate directly to the job described in the JD Context below**. Tailor questions to the stated tech stack, responsibilities, and qualifications.
- If the current topic is **Coding**, at least one of the two questions must require the user to write actual code (e.g., a function implementation or small algorithm). The other may be a code-writing, debugging, or code-explanation question.
- **Evaluate answers based *strictly* on whether they directly answer the specific question asked. If the answer is unrelated, generic, or off-topic—even if technically correct—assign a score no higher than 3.**
- Format responses as **valid JSON only** (no explanations or extra text).

%s

**Current State:**
- You have already covered the following topics: %s
- You are currently on the topic: %s
- This is question number %d out of 2 for this topic

**Topics to Cover in Order:**
1. **Introduction**
2. **Coding**
3. **System Design**
4. **Databases**
5. **Behavioral**
6. **General Backend Knowledge**

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
}`, jdSummary, strings.Join(completedTopics, ", "), currentTopic, questionNumber)
}

func BuildJDPromptInput(jd string) string {
	return fmt.Sprintf(`Your task is to break the following job description into structured JSON under three categories:

- Domain
- Responsibilities
- Qualifications
- Tech Stack

**Before anything else**, first carefully scan the entire job description—including paragraphs and bullets—for any mention of:
- Specific numbers of years of experience
- Platform expertise (e.g., Linux, Windows, embedded systems)
- Low-level systems knowledge
- High-level seniority claims (e.g., “Principal Engineer”, “20+ years”, “technical leadership”)

**Any such statements must be included word-for-word in the qualifications list**, even if they only appear in the introduction or narrative sections. This is critical.

Then extract all other relevant items under each category. Include high-signal technical terms, tools, and processes.

Finally, based on the qualifications and overall language in the job description, classify the expected experience level as one of the following: "junior", "mid-level", or "senior".

Return only **valid JSON** in the following format:

{
  "domain": "...",
  "responsibilities": ["..."],
  "qualifications": ["..."],
  "tech_stack": ["..."]
  "level": "junior, mid-level, or senior"
}

Job description:
%s`, jd)
}

func BuildJDPromptSummary(jdSummary string) string {
	return fmt.Sprintf(`Your task is to extract the following structured fields from the job description below:

- "responsibilities": Select the 3 most technically demanding responsibilities.
- "qualifications": Select the 3 most technically demanding qualifications.
- "tech_stack": Extract the complete list of technologies and tools mentioned in the job.
- "domain": Infer a short descriptive phrase summarizing the product or industry domain (e.g., “HR systems for startups”, “developer infrastructure”, “fintech compliance tooling”).

When selecting the most technically demanding items, prioritize:

- System-level design and architecture
- Debugging and memory management
- Low-level systems (especially Linux)
- Security, observability, identity
- Leadership and mentorship
- Automation, CI/CD, and workflow tools
- Any qualification indicating exceptional seniority or experience (e.g., 20+ years)

Then, based on the qualifications and overall language in the job description, classify the expected experience level as one of the following: "junior", "mid-level", or "senior".

Return only **valid JSON** in the following format:

{
  "domain": "short descriptive phrase",
  "responsibilities": ["..."],
  "qualifications": ["..."],
  "tech_stack": ["..."],
  "level": "junior, mid-level, or senior"
}

Here is the input data:
%s`, jdSummary)
}

type AIClient interface {
	GetChatGPTResponse(prompt string) (*ChatGPTResponse, error)
	GetChatGPTResponseConversation(conversationHistory []map[string]string) (*ChatGPTResponse, error)
	GetChatGPT35Response(prompt string) (*ChatGPTResponse, error)
	ExtractJDInput(jd string) (*JDParsedOutput, error)
	ExtractJDSummary(jdInput *JDParsedOutput) (string, error)
}
