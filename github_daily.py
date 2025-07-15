import json
import os
from datetime import datetime, timedelta
from http import HTTPStatus

import requests
from dotenv import load_dotenv
from openai import OpenAI

load_dotenv()


class GitHubDailyReporter:
    def __init__(self):
        self.github_user = os.getenv("GH_USER")
        self.github_token = os.getenv("GH_TOKEN")
        self.discord_webhook = os.getenv("DISCORD_WEBHOOK_URL")

    def get_github_events(self):
        """Captura eventos do GitHub das últimas 24 horas (públicos e privados)."""
        headers = {
            "Authorization": f"Bearer {self.github_token}",
            "Accept": "application/vnd.github.v3+json",
            "X-GitHub-Api-Version": "2022-11-28"
        }
        
        # Eventos do usuário autenticado (inclui privados)
        url = f"https://api.github.com/users/{self.github_user}/events"
        response = requests.get(url, headers=headers)
        response.raise_for_status()
        
        if response.status_code != HTTPStatus.OK:
            print(f"Erro ao buscar eventos do usuário: {response.status_code}")
            return []
        
        all_events = response.json()

        yesterday = datetime.now() - timedelta(hours=24)
        recent_events = []

        for event in all_events:
            try:
                event_date = datetime.strptime(
                    event["created_at"], "%Y-%m-%dT%H:%M:%SZ"
                )
                if event_date >= yesterday:
                    recent_events.append(event)
            except (KeyError, ValueError) as e:
                print(f"Erro ao processar evento: {e}")
                continue

        recent_events.sort(key=lambda x: x["created_at"], reverse=True)
        return recent_events

    def format_events(self, events):
        """Formata eventos para análise"""
        formatted_events = []

        for event in events:
            event_info = {
                "type": event["type"],
                "repo": event["repo"]["name"],
                "created_at": event["created_at"],
                "is_private": event.get("public", True) == False
            }

            if event["type"] == "PushEvent":
                event_info["commits"] = len(event["payload"]["commits"])
                event_info["branch"] = event["payload"]["ref"].replace(
                    "refs/heads/", ""
                )
                if event["payload"]["commits"]:
                    event_info["commit_messages"] = [
                        commit["message"] for commit in event["payload"]["commits"]
                    ]
            elif event["type"] == "CreateEvent":
                event_info["ref_type"] = event["payload"]["ref_type"]
                if "ref" in event["payload"]:
                    event_info["ref"] = event["payload"]["ref"]
            elif event["type"] == "IssuesEvent":
                event_info["action"] = event["payload"]["action"]
            elif event["type"] == "PullRequestEvent":
                event_info["action"] = event["payload"]["action"]
                event_info["pr_title"] = event["payload"]["pull_request"]["title"]
            elif event["type"] == "DeleteEvent":
                event_info["ref_type"] = event["payload"]["ref_type"]
                event_info["ref"] = event["payload"]["ref"]

            formatted_events.append(event_info)

        return formatted_events

    def generate_twitter_message(self, events):
        """Gera mensagem para Twitter usando OpenAI"""
        if not events:
            return "Hoje foi um dia de planejamento e reflexão no código."

        events_summary = json.dumps(events, indent=2)

        prompt = f"""
            Você é um assistente que recebe as atividades feitas no GitHub hoje, incluindo ações em
            repositórios privados. Com base nelas, gere um breve resumo em texto corrido:

            - Sem emojis  
            - Sem hashtags  
            - Nada clichê ou genérico  
            - Max 280 caracteres

            Atividades do dia:
            {events_summary}
        """.strip()


        try:
            client = OpenAI(api_key=os.getenv("OPENAI_API_KEY"))

            response = client.chat.completions.create(
                model="gpt-4o-mini",
                messages=[
                    {
                        "role": "system",
                        "content": "Você é um desenvolvedor experiente criando textos sobre sua atividade de programação.",
                    },
                    {"role": "user", "content": prompt},
                ],
                max_tokens=150,
                temperature=1.0,
            )

            return response.choices[0].message.content.strip()
        except Exception as e:
            print(f"Erro na OpenAI: {str(e)}")
            return f"Trabalhando em projetos interessantes hoje! {len(events)} atividades no GitHub"

    def send_to_discord(self, message):
        """Envia mensagem para Discord via webhook"""
        embed = {
            "title": "📅 GitHub Daily",
            "description": message,
            "color": 0x7289DA,
            "timestamp": datetime.now().isoformat(),
            "footer": {"text": "GitHub Daily Reporter"},
        }

        payload = {"embeds": [embed]}

        response = requests.post(self.discord_webhook, json=payload)

        if response.status_code != HTTPStatus.NO_CONTENT:
            raise Exception(
                f"Erro ao enviar para Discord: {response.status_code}"
            )

        return True

    def run(self):
        """Executa o processo completo"""
        try:
            print("Buscando eventos do GitHub...")
            events = self.get_github_events()
            formatted_events = self.format_events(events)

            print(f"Encontrados {len(formatted_events)} eventos")

            print("Gerando mensagem com OpenAI...")
            twitter_message = self.generate_twitter_message(formatted_events)

            print("Enviando para Discord...")
            self.send_to_discord(twitter_message)

            print("Processo concluído com sucesso!")
            print(f"Mensagem gerada: {twitter_message}")

        except Exception as e:
            error_message = f"Erro no processo: {str(e)}"
            print(error_message)

            # Tentar enviar erro para Discord
            try:
                self.send_to_discord(
                    f"Erro no GitHub Daily Reporter: {str(e)}"
                )
            except Exception as e:
                print("Erro ao executar o envio da mensagem ao Discord")


if __name__ == "__main__":
    reporter = GitHubDailyReporter()
    reporter.run()
