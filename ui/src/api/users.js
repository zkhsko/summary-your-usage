async function request(path = '', options = {}) {
  let response
  try {
    response = await fetch(`/api/v1/users${path}`, {
      ...options,
      headers: options.body ? { 'Content-Type': 'application/json' } : undefined,
    })
  } catch {
    throw new Error('网络连接失败，请稍后重试')
  }
  if (!response.ok) {
    const body = await response.json().catch(() => null)
    throw new Error(body?.error?.message || '请求失败，请稍后重试')
  }
  if (response.status !== 204) return response.json()
}

export const usersApi = {
  list: () => request(),
  create: input => request('', { method: 'POST', body: JSON.stringify(input) }),
  update: (id, input) => request(`/${id}`, { method: 'PUT', body: JSON.stringify(input) }),
  delete: id => request(`/${id}`, { method: 'DELETE' }),
}
