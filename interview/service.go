package interview

import (
	"log"
	"time"

	"github.com/michaelboegner/interviewer/chatgpt"
	"github.com/michaelboegner/interviewer/user"
)

func StartInterview(repo InterviewRepo, ai chatgpt.AIClient, user *user.User, length, numberQuestions int, difficulty string) (*Interview, error) {

	now := time.Now().UTC()
	prompt := "You are conducting a structured backend development interview. " +
		"The interview follows **six topics in this order**:\n\n" +
		"1. **Introduction**\n" +
		"2. **Coding**\n" +
		"3. **System Design**\n" +
		"4. **Databases**\n" +
		"5. **Behavioral**\n" +
		"6. **General Backend Knowledge**\n\n" +
		"You have already covered the following topics: [].\n" +
		"You are currently on the topic: Introduction. \n\n" +
		"**Rules:**\n" +
		"- Ask **exactly 2 questions per topic** before moving to the next.\n" +
		"- Do **not** skip or reorder topics.\n" +
		"- You only have access to the current topic’s conversation history. Infer progression logically.\n" +
		"- Format responses as **valid JSON only** (no explanations or extra text).\n\n" +
		"**If candidate says 'I don't know':**\n" +
		"- Assign **score: 1** and provide minimal feedback.\n" +
		"- Move to the next question.\n\n" +
		"**JSON Response Format:**\n" +
		"{\n" +
		"    \"topic\": \"current topic\",\n" +
		"    \"subtopic\": \"current subtopic\",\n" +
		"    \"question\": \"previous question\",\n" +
		"    \"score\": the score (1-10) you think the previous answer deserves, default to 0 if you don't have a score,\n" +
		"    \"feedback\": \"brief feedback\",\n" +
		"    \"next_question\": \"next question\",\n" +
		"    \"next_topic\": \"next topic\",\n" +
		"    \"next_subtopic\": \"next subtopic\"\n" +
		"}"

	chatGPTResponse, err := ai.GetChatGPTResponseInterview(prompt)
	if err != nil {
		log.Printf("getChatGPTResponse err: %v\n", err)
		return nil, err
	}
	chatGPTResponse.CreatedAt = now

	interview := &Interview{
		UserId:          user.ID,
		Length:          length,
		NumberQuestions: numberQuestions,
		Difficulty:      difficulty,
		Status:          "Running",
		Score:           100,
		Language:        "Python",
		Prompt:          prompt,
		ChatGPTResponse: chatGPTResponse,
		FirstQuestion:   chatGPTResponse.NextQuestion,
		Subtopic:        chatGPTResponse.Subtopic,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	id, err := repo.CreateInterview(interview)
	if err != nil {
		log.Printf("CreateInterview err: %v", err)
		return nil, err
	}
	interview.Id = id

	return interview, nil
}

func GetInterview(repo InterviewRepo, interviewID int) (*Interview, error) {
	interview, err := repo.GetInterview(interviewID)
	if err != nil {
		return nil, err
	}

	return interview, nil
}
