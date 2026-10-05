import {
  ActionIcon,
  AppShell,
  Burger,
  Button,
  Group,
  NavLink as MantineNavLink,
  ScrollArea,
  Text,
  Tooltip,
  useComputedColorScheme,
  useMantineColorScheme,
} from '@mantine/core'
import { useDisclosure } from '@mantine/hooks'
import {
  IconAdjustments,
  IconChartLine,
  IconCoin,
  IconHome,
  IconListDetails,
  IconLogout,
  IconMoon,
  IconReportAnalytics,
  IconSun,
  IconTimeline,
  IconTrendingUp,
  IconWallet,
} from '@tabler/icons-react'
import type { Icon } from '@tabler/icons-react'
import { Outlet, useLocation, useNavigate } from 'react-router'

import { useLogout } from '@/api/auth'

type NavItem = {
  label: string
  to: string
  icon: Icon
}

const navItems: NavItem[] = [
  { label: 'Inicio', to: '/', icon: IconHome },
  { label: 'Activos', to: '/activos', icon: IconWallet },
  { label: 'Transaccións', to: '/transaccions', icon: IconListDetails },
  { label: 'Resultados', to: '/resultados', icon: IconTrendingUp },
  { label: 'Dividendos', to: '/dividendos', icon: IconCoin },
  { label: 'Informes', to: '/informes', icon: IconReportAnalytics },
  { label: 'Gráficas', to: '/graficas', icon: IconChartLine },
  { label: 'Proxección', to: '/proxeccion', icon: IconTimeline },
  { label: 'Axustes', to: '/axustes', icon: IconAdjustments },
]

export function AppLayout() {
  const [opened, { toggle, close }] = useDisclosure()
  const location = useLocation()
  const navigate = useNavigate()
  const logout = useLogout()
  const { setColorScheme } = useMantineColorScheme()
  const computedColorScheme = useComputedColorScheme('light', { getInitialValueInEffect: true })

  const handleLogout = () => {
    logout.mutate(undefined, {
      onSettled: () => navigate('/login', { replace: true }),
    })
  }

  return (
    <AppShell
      header={{ height: 64 }}
      navbar={{ width: 280, breakpoint: 'sm', collapsed: { mobile: !opened } }}
      padding="md"
    >
      <AppShell.Header>
        <Group h="100%" px="md" justify="space-between" wrap="nowrap">
          <Group gap="sm" wrap="nowrap" style={{ minWidth: 0 }}>
            <Burger opened={opened} onClick={toggle} hiddenFrom="sm" size="sm" aria-label="Abrir navegación" />
            <Text fw={700} size="lg" truncate="end">
              Control de Investimentos
            </Text>
          </Group>
          <Group gap="xs" wrap="nowrap">
            <Tooltip label={computedColorScheme === 'dark' ? 'Modo claro' : 'Modo escuro'}>
              <ActionIcon
                variant="subtle"
                color="gray"
                onClick={() => setColorScheme(computedColorScheme === 'dark' ? 'light' : 'dark')}
                aria-label="Cambiar esquema de cor"
              >
                {computedColorScheme === 'dark' ? <IconSun size={20} /> : <IconMoon size={20} />}
              </ActionIcon>
            </Tooltip>
            <Button
              variant="light"
              leftSection={<IconLogout size={16} />}
              onClick={handleLogout}
              loading={logout.isPending}
              visibleFrom="sm"
            >
              Saír
            </Button>
            <Tooltip label="Saír">
              <ActionIcon
                variant="light"
                onClick={handleLogout}
                loading={logout.isPending}
                aria-label="Saír"
                hiddenFrom="sm"
              >
                <IconLogout size={18} />
              </ActionIcon>
            </Tooltip>
          </Group>
        </Group>
      </AppShell.Header>

      <AppShell.Navbar p="md">
        <ScrollArea>
          {navItems.map((item) => {
            const IconComponent = item.icon
            const active = item.to === '/' ? location.pathname === '/' : location.pathname.startsWith(item.to)
            return (
              <MantineNavLink
                key={item.to}
                label={item.label}
                leftSection={<IconComponent size={18} />}
                active={active}
                onClick={() => {
                  navigate(item.to)
                  close()
                }}
                style={{ borderRadius: 'var(--mantine-radius-md)' }}
              />
            )
          })}
        </ScrollArea>
      </AppShell.Navbar>

      <AppShell.Main>
        <Outlet />
      </AppShell.Main>
    </AppShell>
  )
}
