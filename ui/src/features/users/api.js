import { request } from '../../framework/http'

export const usersApi = {
  list: () => request('/users'),
  create: input => request('/users', { method: 'POST', body: JSON.stringify(input) }),
  update: (id, input) => request(`/users/${id}`, { method: 'PUT', body: JSON.stringify(input) }),
  delete: id => request(`/users/${id}`, { method: 'DELETE' }),
}
