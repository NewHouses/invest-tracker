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
  IconArrowsExchange,
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

type NavLeaf = {
  label: string
  to: string
  icon: Icon
}

type NavGroup = {
  label: string
  icon: Icon
  children: NavLeaf[]
}

type NavItem = NavLeaf | NavGroup

const navItems: NavItem[] = [
  { label: 'Inicio', to: '/', icon: IconHome },
  { label: 'Portofolio', to: '/activos', icon: IconWallet },
  { label: 'Dividendos', to: '/dividendos', icon: IconCoin },
  {
    label: 'Operacións',
    icon: IconArrowsExchange,
    children: [
      { label: 'Transaccións', to: '/transaccions', icon: IconListDetails },
      { label: 'Resultados', to: '/resultados', icon: IconTrendingUp },
    ],
  },
  { label: 'Informes', to: '/informes', icon: IconReportAnalytics },
  { label: 'Gráficas', to: '/graficas', icon: IconChartLine },
  { label: 'Proxeccións', to: '/proxeccion', icon: IconTimeline },
  { label: 'Axustes', to: '/axustes', icon: IconAdjustments },
]

function isActive(pathname: string, to: string) {
  return to === '/' ? pathname === '/' : pathname === to || pathname.startsWith(`${to}/`)
}

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
            const linkStyle = { borderRadius: 'var(--mantine-radius-md)' }
            if ('children' in item) {
              return (
                <MantineNavLink
                  key={item.label}
                  label={item.label}
                  leftSection={<IconComponent size={18} />}
                  defaultOpened
                  childrenOffset={28}
                  style={linkStyle}
                >
                  {item.children.map((child) => {
                    const ChildIcon = child.icon
                    return (
                      <MantineNavLink
                        key={child.to}
                        label={child.label}
                        leftSection={<ChildIcon size={16} />}
                        active={isActive(location.pathname, child.to)}
                        onClick={() => {
                          navigate(child.to)
                          close()
                        }}
                        style={linkStyle}
                      />
                    )
                  })}
                </MantineNavLink>
              )
            }
            return (
              <MantineNavLink
                key={item.to}
                label={item.label}
                leftSection={<IconComponent size={18} />}
                active={isActive(location.pathname, item.to)}
                onClick={() => {
                  navigate(item.to)
                  close()
                }}
                style={linkStyle}
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
