export const translateError = (err: unknown): string => {
    const errorObj = err as { error?: string, statusText?: string, message?: string };
    const code = errorObj?.error || errorObj?.statusText?.toLowerCase() || '';
    const message = errorObj?.message || '';

    const translations: Record<string, string> = {
        'invalid_credentials': 'Ungültige Anmeldedaten.',
        'unauthorized': 'Ungültige Anmeldedaten.',
        'email_exists': 'Diese E-Mail-Adresse ist bereits registriert.',
        'user_not_found': 'Benutzer nicht gefunden.',
        'email_not_verified': message || 'Bitte bestätige erst deine E-Mail-Adresse.',
        'invalid_request_body': 'Ungültige Anfrage. Bitte fülle alle Felder korrekt aus.',
        'server_error': 'Serverfehler. Bitte versuche es später noch einmal.',
        'database_error': 'Datenbankfehler. Bitte versuche es später noch einmal.',
        'invalid_csrf': 'Sitzung abgelaufen. Bitte lade die Seite neu.',
        'invalid_token': 'Der Bestätigungslink ist ungültig oder abgelaufen.',
        'account_disabled': 'Dein Account wurde deaktiviert.',
        'invalid_input': 'Ungültige Eingabe. Bitte prüfe deine Angaben.'
    };

    if (translations[code]) return translations[code];

    const msg = message.toLowerCase();
    if (msg.includes('invalid email or password')) return 'Ungültige Anmeldedaten.';
    if (msg.includes('already exists')) return 'Diese E-Mail-Adresse ist bereits registriert.';
    if (msg.includes('failed to fetch')) return 'Verbindung zum Server fehlgeschlagen. Bitte prüfe deine Internetverbindung.';
    if (msg.includes('network error')) return 'Netzwerkfehler. Bitte prüfe deine Verbindung.';
    if (msg.includes('timeout')) return 'Zeitüberschreitung. Der Server antwortet nicht.';

    return message || 'Ein unerwarteter Fehler ist aufgetreten.';
};
