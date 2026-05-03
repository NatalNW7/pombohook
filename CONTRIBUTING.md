# Contribuindo para o PomboHook

Obrigado por se interessar em contribuir com o **PomboHook**! 🎉

Este documento serve como um guia para ajudar você a entender o nosso processo de desenvolvimento, os padrões de código exigidos e como você pode enviar suas alterações.

## 🤝 Processo de Submissão

Antes de começar a escrever código para uma nova funcionalidade (feature) ou para a correção de um bug, **é obrigatório abrir uma Issue**.

1. **Abra uma Issue:** Vá até a aba "Issues" no GitHub e descreva detalhadamente o que você deseja implementar ou qual problema encontrou.
2. **Discussão:** Aguarde a validação da comunidade ou do mantenedor. Isso evita que você perca tempo implementando algo que possa não se alinhar com a visão do projeto ou que já esteja sendo feito por outra pessoa.
3. **Desenvolvimento:** Após a aprovação da Issue, você pode fazer o fork do repositório, criar a sua branch e começar a codar.
4. **Pull Request:** Abra o seu PR referenciando a Issue original (ex: `Resolves #12`).

## 💻 Padrões de Código e Engenharia

Para manter o projeto sustentável e organizado, seguimos regras rigorosas de engenharia. Ao contribuir, por favor, certifique-se de seguir estas diretrizes:

1. **Simplicidade:** Prefira funções e arquivos pequenos. Evite aninhamentos profundos (use *early returns* e *guard clauses*).
2. **Responsabilidade Única (SRP):** Cada módulo, classe e função deve ter apenas uma responsabilidade.
3. **Nomenclatura:** Use nomes que revelem intenção (evite nomes genéricos como `data`, `process` ou `handler`).
4. **Comentários:** Não comente o que é óbvio. Use comentários apenas para explicar decisões não triviais ou contornos de bugs.
5. **Erros Acionáveis:** As mensagens de erro e exceções devem incluir contexto suficiente para identificar o problema rapidamente. Não use mensagens vagas.
6. **Logging:** Use logging estruturado para observabilidade e logs em texto plano para saídas voltadas ao usuário no CLI.

## 🧪 Fluxo de Testes

O PomboHook leva testes muito a sério. Para que o seu Pull Request seja aceito, os seguintes requisitos de teste **devem ser atendidos rigorosamente**:

1. **Execução Local:** Os comandos `make test` e `make lint` (ou `go vet ./...`) devem passar sem nenhum erro na sua máquina.
2. **TDD e Comportamento:** Teste o comportamento público, *edge cases* e os caminhos de falha.
3. **Cobertura:** 
   - Códigos novos exigem **100% de cobertura** das novas linhas.
   - O projeto possui um piso global de **80% de cobertura**, e módulos críticos (auth, rotas, túnel) exigem pelo menos **90%**.
4. **Mocking:** Não faça mock da unidade que está sendo testada, faça mock apenas de suas dependências.

## 📝 Padrão de Commits

Nós adotamos a convenção do [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/). Suas mensagens de commit devem seguir este formato padrão:

```text
<tipo>[escopo opcional]: <descrição>
```

**Exemplos válidos:**
- `feat: add persistent sqlite storage`
- `fix: resolve data race in tunnel manager`
- `docs: update setup instructions in README`
- `test: improve coverage for RunSleep`

## 💬 Comunidade e Dúvidas

Se você tiver qualquer dúvida sobre a arquitetura do código, como configurar seu ambiente ou como abordar a resolução de uma Issue, não hesite em perguntar! 

O nosso principal canal de comunicação são as **próprias Issues do GitHub**. Sinta-se à vontade para comentar e marcar o mantenedor para pedir ajuda.
