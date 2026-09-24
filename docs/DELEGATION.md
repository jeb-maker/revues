# Délégation — GitHub

1. Lire [AGENTS.md](../AGENTS.md) (contrat agents)
2. Une issue = une PR · branche `cursor/issue-<N>-<slug>-f21b`
3. `./scripts/check.sh` vert avant push · corps PR : `Closes #<N>`
4. Pièges : [REVIEW_ADVERSE.md](./REVIEW_ADVERSE.md) · DoD : [DEFINITION_OF_DONE.md](./DEFINITION_OF_DONE.md)

```
Implémente UNIQUEMENT l'issue #N du repo jeb-maker/revues.
Lis AGENTS.md. ./scripts/check.sh avant push. PR : Closes #N.
```

Revue humaine obligatoire : OAuth, `area:integrations`, chiffrement, webhooks.

Labels usuels : `area:infra|data|auth|core|ui|admin|integrations|notifications|attachments`, `epic`.

Réf. : [PLAN.md](./PLAN.md) · [RBAC.md](./RBAC.md) · [Issues](https://github.com/jeb-maker/revues/issues)
