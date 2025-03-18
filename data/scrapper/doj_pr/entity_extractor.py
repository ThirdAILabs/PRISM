import time
from openai import OpenAI


def ask_gpt(message):
    client = OpenAI()
    while True:
        try:
            response = client.chat.completions.create(
                model="gpt-4o-mini",
                messages=[{"role": "user", "content": message}],
                temperature=0.1,
            )
            return response.choices[0].message.content
        except Exception as e:
            time.sleep(5)


def get_entities(content):
    prompt = f"In the following article, list all NON-US INSTITUTIONS responsible for the crime. Please DO NOT list the names of countries or cities. DO NOT use abbreviations. If there is not a crime, list nothing. Please return a list of all of these CRIMINAL institutions in a newline separated list such that I can parse the response with response.split('\\n') in python. Include no special characters like '`'. Here is the article for you to examine: {content}"
    institutions = ask_gpt(prompt)
    prompt = f"In the following article, list the NAMES of the individual person or persons who are RESPONSIBLE for the crime. Who are the ones at fault, not the ones after them or innocent. If there is not a crime, list nothing. Please return a list of all of these names in a newline separated list such that I can parse the response with response.split('\\n') in python. Include no special characters like '`'. Here is the article for you to examine: {content}"
    persons = ask_gpt(prompt)
    return institutions.split("\n") + persons.split("\n")
