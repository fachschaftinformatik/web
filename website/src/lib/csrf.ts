import { getAuthCsrf, type DtoCsrfResponse } from './api';

let csrfToken: string | null = null;
let fetchPromise: Promise<DtoCsrfResponse | null> | null = null;

export const getCsrfToken = () => csrfToken;

export const setCsrfToken = (token: string) => {
    csrfToken = token;
};

export const clearCsrfToken = () => {
    csrfToken = null;
    fetchPromise = null;
};

export const fetchCsrfToken = async (): Promise<DtoCsrfResponse | null> => {
    if (fetchPromise) return fetchPromise;

    fetchPromise = (async () => {
        try {
            const { data } = await getAuthCsrf();
            if (data?.csrf) {
                csrfToken = data.csrf;
                return data;
            }
        } catch (error) {
            console.error('Failed to fetch CSRF token:', error);
        } finally {
            fetchPromise = null;
        }
        return null;
    })();

    return fetchPromise;
};

export const getCsrfFromCookie = (): string | null => {
    if (typeof document === 'undefined') return null;
    const name1 = '__Host-csrf=';
    const name2 = 'csrf=';
    const decodedCookie = decodeURIComponent(document.cookie);
    const ca = decodedCookie.split(';');
    
    let fallbackToken: string | null = null;

    for (let i = 0; i < ca.length; i++) {
        let c = ca[i];
        while (c.charAt(0) === ' ') c = c.substring(1);
        if (c.indexOf(name1) === 0) return c.substring(name1.length, c.length);
        if (c.indexOf(name2) === 0) fallbackToken = c.substring(name2.length, c.length);
    }
    return fallbackToken;
};
