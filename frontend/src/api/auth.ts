import { useEffect } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { apiFetch, subscribeUnauthorized } from '@/api/client'
import { queryKeys } from '@/api/queryKeys'
import type { AuthenticatedResponse, AuthStatusResponse } from '@/api/types'

export function useAuthStatus() {
  return useQuery({
    queryKey: queryKeys.auth.status,
    queryFn: ({ signal }) => apiFetch<AuthStatusResponse>('/api/auth/status', { signal }),
  })
}

export function useSetup() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (password: string) =>
      apiFetch<AuthenticatedResponse>('/api/auth/setup', { method: 'POST', body: { password } }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.auth.status }),
  })
}

export function useLogin() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: { password: string; remember: boolean }) =>
      apiFetch<AuthenticatedResponse>('/api/auth/login', { method: 'POST', body: input }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.auth.status }),
  })
}

export function useLogout() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => apiFetch<void>('/api/auth/logout', { method: 'POST' }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.auth.status }),
  })
}

export function useChangePassword() {
  return useMutation({
    mutationFn: (input: { currentPassword: string; newPassword: string }) =>
      apiFetch<void>('/api/auth/password', { method: 'POST', body: input }),
  })
}

export function AuthUnauthorizedListener() {
  const queryClient = useQueryClient()

  useEffect(
    () =>
      subscribeUnauthorized(() => {
        void queryClient.invalidateQueries({ queryKey: queryKeys.auth.status })
      }),
    [queryClient],
  )

  return null
}
