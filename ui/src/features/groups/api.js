import { request } from '../../framework/http'

export const groupsApi = {
  list: () => request('/groups'),
  create: input => request('/groups', { method: 'POST', body: JSON.stringify(input) }),
  update: (id, input) => request(`/groups/${id}`, { method: 'PUT', body: JSON.stringify(input) }),
  delete: id => request(`/groups/${id}`, { method: 'DELETE' }),
}
