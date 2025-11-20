// website/src/lib/api.ts

// TODO: Handle interfaces via OpenAPI at some point
export interface User {
  id: string;
  email: string;
  name: string;
  role: 'user' | 'editor' | 'admin';
  active: 0 | 1;
  verified: 0 | 1;
  verified_at?: string | null;
  verified_until?: string | null;
  programid: number;
  created_at: string;
  updated_at: string;
}

export interface Program {
  id: number;
  name: string;
  versions: string[];
}

interface ApiError {
  error: string;
  message: string;
}

async function handleResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    let errorData: ApiError = { error: `http_${response.status}`, message: response.statusText };
    try {
      errorData = await response.json();
    } catch {
      // ignore non-JSON error bodies
    }
    console.error('API Error:', errorData);
    throw new Error(errorData.message || `Request failed with status ${response.status}`);
  }
  if (response.status === 204) {
      return null as T;
  }
  return response.json() as Promise<T>;
}


let csrfToken: string | null = null;
async function getCsrfToken(): Promise<string> {
    if (csrfToken)
        return csrfToken;
    const response = await fetch('/api/auth/csrf');
    const data = await handleResponse<{ csrf: string }>(response);
    csrfToken = data.csrf;
    return csrfToken;
}

export async function registerUser(userData: Omit<User, 'id' | 'role' | 'active' | 'verified' | 'created_at' | 'updated_at' | 'verified_at' | 'verified_until'> & { password: string }): Promise<User> {
  const response = await fetch('/api/auth/register', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(userData),
  });
  return handleResponse<User>(response);
}

export async function loginUser(credentials: { email: string; password: string; rememberMe?: boolean }): Promise<User> {
  const response = await fetch('/api/auth/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(credentials),
  });
  const user = await handleResponse<User>(response);
  csrfToken = null;
  return user;
}

export async function logoutUser(): Promise<void> {
  const token = await getCsrfToken();
  const response = await fetch('/api/auth/logout', {
    method: 'POST',
    headers: {
      'X-CSRF-Token': token,
    },
  });
  csrfToken = null;
  await handleResponse<void>(response);
}

export async function getMe(): Promise<User> {
  const response = await fetch('/api/auth/me');
  return handleResponse<User>(response);
}

export async function getPrograms(): Promise<Program[]> {
  const response = await fetch('/api/programs');
  return handleResponse<Program[]>(response);
}
