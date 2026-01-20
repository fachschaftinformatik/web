import { getAuthCsrf } from './api/sdk.gen';

let csrfToken: string | null = null;

export const getCsrfToken = () => csrfToken;

export const setCsrfToken = (token: string) => {
    csrfToken = token;
};

export const fetchCsrfToken = async () => {
    try {
        const { data } = await getAuthCsrf();
        if (data?.csrf) {
            csrfToken = data.csrf;
            return data.csrf;
        }
    } catch (error) {
        console.error('Failed to fetch CSRF token:', error);
    }
    return null;
};

export const getCsrfFromCookie = (): string | null => {
    if (typeof document === 'undefined') return null;
    const name = '__Host-csrf=';
    const decodedCookie = decodeURIComponent(document.cookie);
    const ca = decodedCookie.split(';');
    for (let i = 0; i < ca.length; i++) {
        let c = ca[i];
        while (c.charAt(0) === ' ') {
            c = c.substring(1);
        }
        if (c.indexOf(name) === 0) {
            return c.substring(name.length, c.length);
        }
    }
    return null;
};
