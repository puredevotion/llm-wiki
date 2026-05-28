const BASE_URL = import.meta.env.VITE_API_URL || "http://localhost:8080/api/v1";

export async function apiFetch<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const response = await fetch(`${BASE_URL}${endpoint}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
  });

  if (!response.ok) {
    throw new Error(`API error: ${response.statusText}`);
  }

  if (response.status === 204) {
    return {} as T;
  }

  return response.json();
}
