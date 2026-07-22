package auth

import "strings"

// LoginErrorMessage returns a user-facing message for login ?error= query values.
func LoginErrorMessage(code string) string {
	switch strings.TrimSpace(code) {
	case "":
		return ""
	case "email non autorisé":
		// Generic copy — no confirmation whether the email exists in any whitelist.
		return "Connexion impossible avec ce compte GitHub. " +
			"Demandez à un administrateur de votre organisation de vous autoriser, " +
			"ou vérifiez que votre email GitHub est vérifié et accessible à l'application OAuth."
	case "oauth non configuré":
		return "La connexion GitHub OAuth n'est pas configurée sur ce serveur. " +
			"L'administrateur doit renseigner REVUES_GITHUB_CLIENT_ID et REVUES_GITHUB_CLIENT_SECRET."
	case "session oauth invalide":
		return "La session de connexion GitHub a expiré. Réessayez en cliquant sur le bouton ci-dessous."
	case "state invalide":
		return "La vérification de sécurité OAuth a échoué. Réessayez la connexion."
	case "code manquant":
		return "GitHub n'a pas renvoyé de code d'autorisation. Réessayez la connexion."
	case "échec oauth":
		return "Échec de l'échange OAuth avec GitHub. Réessayez ou contactez l'administrateur."
	case "profil github":
		return "Impossible de récupérer votre profil GitHub. Vérifiez les autorisations de l'application OAuth."
	default:
		return code
	}
}
