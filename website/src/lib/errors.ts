const ERROR_MESSAGES: Record<string, string> = {
    'invalid_credentials': 'E-Mail oder Passwort ist nicht korrekt.',
    'user_not_found': 'Dieser Account existiert nicht.',
    'email_already_exists': 'Diese E-Mail-Adresse wird bereits verwendet.',
    'unauthorized': 'Bitte melde dich an, um fortzufahren.',
    'forbidden': 'Du hast nicht die erforderlichen Berechtigungen.',
    'invalid_csrf': 'Deine Sitzung ist abgelaufen. Bitte lade die Seite neu.',

    'weak_password': 'Das Passwort ist zu schwach.',
    'invalid_email': 'Bitte gib eine gültige E-Mail-Adresse ein.',

    'file_too_large': 'Die Datei ist zu groß (maximal 50MB erlaubt).',
    'invalid_file_type': 'Dieser Dateityp wird nicht unterstützt.',
    'upload_failed': 'Der Upload ist fehlgeschlagen. Bitte versuche es erneut.',

    'network_error': 'Netzwerkfehler. Bitte überprüfe deine Internetverbindung.',
    'bad_request': 'Ungültige Anfrage.',
    'internal_server_error': 'Ein Serverfehler ist aufgetreten. Bitte versuche es später erneut.',
    'unknown_error': 'Ein unbekannter Fehler ist aufgetreten.',
};

export function getFriendlyErrorMessage(error: unknown): string {
    if (!error) return ERROR_MESSAGES.unknown_error;

    if (typeof error === 'string') {
        return ERROR_MESSAGES[error] || error || ERROR_MESSAGES.unknown_error;
    }

    if (typeof error === 'object' && error !== null) {
        const errObj = error as Record<string, unknown>;
        const errorCode = (errObj.error || errObj.code || errObj.message) as string | undefined;

        if (errorCode && typeof errorCode === 'string' && ERROR_MESSAGES[errorCode]) {
            return ERROR_MESSAGES[errorCode];
        }

        const message = (errObj.message || '') as string;
        if (message.includes('Network Error') || message.includes('Failed to fetch')) {
            return ERROR_MESSAGES.network_error;
        }

        if (message) return message;
    }

    return ERROR_MESSAGES.unknown_error;
}
