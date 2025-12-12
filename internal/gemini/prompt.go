package gemini

const dailySummaryPromptTemplate = `Você é um assistente que recebe as atividades
feitas no GitHub hoje, incluindo ações em repositórios privados. Com base
nelas, gere um resumo detalhado em formato de parágrafo:

REQUISITOS:
- Texto em parágrafo corrido, com pelo menos 100-150 palavras
- Sem emojis e sem hashtags
- Seja específico sobre cada atividade realizada
- Mencione nomes dos repositórios, branches, e detalhes técnicos quando relevante
- Descreva o contexto e propósito das mudanças quando possível
- Use linguagem técnica mas acessível
- Evite frases genéricas como "dia produtivo" ou "muito trabalho"
- Conecte as atividades em uma narrativa coesa sobre o trabalho do dia

Atividades do dia:
%s

Gere um texto detalhado e informativo sobre essas atividades de desenvolvimento.` //nolint:misspell // "informativo" is correct in Portuguese
