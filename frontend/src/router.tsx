import { Center, Loader, Stack, Text, Title } from '@mantine/core'
import { lazy, Suspense } from 'react'
import type { ReactNode } from 'react'
import { createBrowserRouter, Navigate, Outlet, useLocation } from 'react-router'

import { useAuthStatus } from '@/api/auth'
import { AppLayout } from '@/components/layout/AppLayout'
import { ConfigurarPage } from '@/pages/auth/ConfigurarPage'
import { LoginPage } from '@/pages/auth/LoginPage'

// As seccións cárganse baixo demanda para que gráficas e informes (Recharts)
// non engorden o paquete inicial.
const InicioPage = lazy(() => import('@/pages/inicio/InicioPage').then((m) => ({ default: m.InicioPage })))
const ActivosPage = lazy(() => import('@/pages/activos/ActivosPage').then((m) => ({ default: m.ActivosPage })))
const ActivoDetailPage = lazy(() =>
  import('@/pages/activos/ActivoDetailPage').then((m) => ({ default: m.ActivoDetailPage })),
)
const TransaccionsPage = lazy(() =>
  import('@/pages/transaccions/TransaccionsPage').then((m) => ({ default: m.TransaccionsPage })),
)
const ResultadosPage = lazy(() =>
  import('@/pages/resultados/ResultadosPage').then((m) => ({ default: m.ResultadosPage })),
)
const DividendosPage = lazy(() =>
  import('@/pages/dividendos/DividendosPage').then((m) => ({ default: m.DividendosPage })),
)
const InformesPage = lazy(() => import('@/pages/informes/InformesPage').then((m) => ({ default: m.InformesPage })))
const GraficasPage = lazy(() => import('@/pages/graficas/GraficasPage').then((m) => ({ default: m.GraficasPage })))
const ProxeccionPage = lazy(() =>
  import('@/pages/proxeccion/ProxeccionPage').then((m) => ({ default: m.ProxeccionPage })),
)
const AxustesPage = lazy(() => import('@/pages/axustes/AxustesPage').then((m) => ({ default: m.AxustesPage })))

function LoadingScreen() {
  return (
    <Center h="60vh">
      <Loader aria-label="Cargando" />
    </Center>
  )
}

function page(element: ReactNode) {
  return <Suspense fallback={<LoadingScreen />}>{element}</Suspense>
}

export function RequireAuth() {
  const location = useLocation()
  const status = useAuthStatus()

  if (status.isLoading) return <LoadingScreen />
  if (status.data?.setupRequired) return <Navigate to="/configurar" replace />
  if (!status.data?.authenticated) {
    const next = encodeURIComponent(`${location.pathname}${location.search}`)
    return <Navigate to={`/login?next=${next}`} replace />
  }
  return <Outlet />
}

export function NotFoundPage() {
  return (
    <Center h="60vh">
      <Stack align="center" gap="xs">
        <Title order={1}>404</Title>
        <Text c="dimmed">Non atopamos esta páxina.</Text>
      </Stack>
    </Center>
  )
}

export const router = createBrowserRouter([
  { path: '/login', element: <LoginPage /> },
  { path: '/configurar', element: <ConfigurarPage /> },
  {
    element: <RequireAuth />,
    children: [
      {
        element: <AppLayout />,
        children: [
          { index: true, element: page(<InicioPage />) },
          { path: 'activos', element: page(<ActivosPage />) },
          { path: 'activos/:id', element: page(<ActivoDetailPage />) },
          { path: 'transaccions', element: page(<TransaccionsPage />) },
          { path: 'resultados', element: page(<ResultadosPage />) },
          { path: 'dividendos', element: page(<DividendosPage />) },
          { path: 'informes', element: page(<InformesPage />) },
          { path: 'graficas', element: page(<GraficasPage />) },
          { path: 'proxeccion', element: page(<ProxeccionPage />) },
          { path: 'axustes', element: page(<AxustesPage />) },
        ],
      },
    ],
  },
  { path: '*', element: <NotFoundPage /> },
])
