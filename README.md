# StoryGo

StoryGo is a creative writing assistant powered by AI. It's designed to help writers, storytellers, and game masters develop their ideas, overcome writer's block, and accelerate the creative process. With a simple and intuitive interface, StoryGo makes it easy to generate, refine, and export your stories.

## Why Use StoryGo?

- **Endless Inspiration:** Generate new ideas, characters, and plot twists when you're feeling stuck.
- **Rapid Development:** Quickly create a detailed overview of your story, including a premise and character descriptions.
- **Creative Collaboration:** Use the AI as a brainstorming partner to explore different narrative paths and possibilities.
- **Image-to-Story:** Use images as a source of inspiration to generate unique story concepts.
- **Simple and Focused:** A minimal interface that lets you focus on your writing.

## Features

- **AI-Powered Generation:** Create rich and detailed story content based on your context and direction, one paragraph at a time.
- **Story Overview Mode:** Generate a high-level overview of your story, including premise and character descriptions.
- **AI-Powered Suggestions:** Get suggestions for the next step in your story.
- **Context and Direction:** Guide the AI with a "World Bible" (context) and a "Next Step" (direction) to maintain consistency and control the narrative.
- **Model Switching:** Choose from a variety of AI models to find the perfect one for your writing style.
- **Vision-Based Inspiration:** Upload an image and let the AI generate a story based on it. Non-vision models are supported by converting the image to a detailed text description.
- **Export to PDF:** Save your stories as PDF files for easy sharing and printing.
- **Keybinds for Efficiency:**
  - `Ctrl+Enter`: Generate/Suggest.
  - `Tab`: Continue inline.
  - `Ctrl+S`: Save as PDF.

## Installation

To get StoryGo up and running, you'll need to have Go installed on your system.

1. **Clone the repository:**
```bash
git clone https://github.com/your-username/storygo.git
cd storygo
```

2. **Install dependencies:**
```bash
go mod tidy
```

3. **Set up your environment variables:**
  - Rename the `.example.env` file to `.env`.
  - Open the `.env` file and add your OpenRouter API key:
```
OPENROUTER_TOKEN="your-openrouter-api-key"
```

## Running the Application

Once you've completed the installation steps, you can run the application with the following command:

```bash
go run .
```

The application will be available at `http://localhost:3344`.
