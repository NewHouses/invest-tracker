import { MantineProvider, localStorageColorSchemeManager } from '@mantine/core'
import { ModalsProvider } from '@mantine/modals'
import { Notifications } from '@mantine/notifications'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { RouterProvider } from 'react-router'

import '@mantine/core/styles.css'
import '@mantine/notifications/styles.css'
import '@mantine/dates/styles.css'
import '@mantine/charts/styles.css'

import { ApiError } from '@/api/client'
import { LanguageProvider } from '@/i18n'
import { router } from '@/router'
import { theme } from '@/theme'

const colorSchemeManager = localStorageColorSchemeManager({ key: 'invest-tracker-color-scheme' })

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: (failureCount, error) => {
        if (error instanceof ApiError && error.status >= 400 && error.status < 500) {
          return false
        }
        return failureCount < 2
      },
    },
    mutations: {
      retry: false,
    },
  },
})

export function App() {
  return (
    <MantineProvider theme={theme} defaultColorScheme="auto" colorSchemeManager={colorSchemeManager}>
      <QueryClientProvider client={queryClient}>
        <LanguageProvider>
          <ModalsProvider>
            <Notifications position="top-right" />
            <RouterProvider router={router} />
          </ModalsProvider>
        </LanguageProvider>
      </QueryClientProvider>
    </MantineProvider>
  )
}
