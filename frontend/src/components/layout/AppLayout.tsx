import {
  ActionIcon,
  AppShell,
  Burger,
  Button,
  Group,
  Menu,
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
  IconCheck,
  IconCoin,
  IconHome,
  IconLanguage,
  IconListDetails,
  IconMoon,
  IconReportAnalytics,
  IconSun,
  IconTimeline,
  IconTrendingUp,
  IconWallet,
} from '@tabler/icons-react'
import type { Icon } from '@tabler/icons-react'
import { Outlet, useLocation, useNavigate } from 'react-router'

import { languages, t, tKey, useLanguage } from '@/i18n'

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

// As etiquetas tradúcense ao debuxar o menú, co idioma de cada momento.
const navItems: NavItem[] = [
  { label: tKey('Inicio'), to: '/', icon: IconHome },
  { label: tKey('Portofolio'), to: '/activos', icon: IconWallet },
  { label: tKey('Dividendos'), to: '/dividendos', icon: IconCoin },
  {
    label: tKey('Operacións'),
    icon: IconArrowsExchange,
    children: [
      { label: tKey('Transaccións'), to: '/transaccions', icon: IconListDetails },
      { label: tKey('Resultados'), to: '/resultados', icon: IconTrendingUp },
    ],
  },
  { label: tKey('Informes'), to: '/informes', icon: IconReportAnalytics },
  { label: tKey('Gráficas'), to: '/graficas', icon: IconChartLine },
  { label: tKey('Proxeccións'), to: '/proxeccion', icon: IconTimeline },
  { label: tKey('Axustes'), to: '/axustes', icon: IconAdjustments },
]

function isActive(pathname: string, to: string) {
  return to === '/' ? pathname === '/' : pathname === to || pathname.startsWith(`${to}/`)
}

function LanguageMenu() {
  const { language, setLanguage } = useLanguage()
  return (
    <Menu position="bottom-end" shadow="md">
      <Menu.Target>
        <Button
          variant="subtle"
          color="gray"
          size="compact-md"
          leftSection={<IconLanguage size={18} />}
          aria-label={t('Idioma')}
          title={t('Idioma')}
        >
          {language.toUpperCase()}
        </Button>
      </Menu.Target>
      <Menu.Dropdown>
        {languages.map((option) => (
          <Menu.Item
            key={option.value}
            lang={option.value}
            onClick={() => setLanguage(option.value)}
            rightSection={option.value === language ? <IconCheck size={16} /> : null}
          >
            {option.label}
          </Menu.Item>
        ))}
      </Menu.Dropdown>
    </Menu>
  )
}

export function AppLayout() {
  const [opened, { toggle, close }] = useDisclosure()
  const location = useLocation()
  const navigate = useNavigate()
  const { setColorScheme } = useMantineColorScheme()
  const computedColorScheme = useComputedColorScheme('light', { getInitialValueInEffect: true })

  return (
    <AppShell
      header={{ height: 64 }}
      navbar={{ width: 280, breakpoint: 'sm', collapsed: { mobile: !opened } }}
      padding="md"
    >
      <AppShell.Header>
        <Group h="100%" px="md" justify="space-between" wrap="nowrap">
          <Group gap="sm" wrap="nowrap" style={{ minWidth: 0 }}>
            <Burger opened={opened} onClick={toggle} hiddenFrom="sm" size="sm" aria-label={t('Abrir navegación')} />
            <Text fw={700} size="lg" truncate="end">
              Control de Investimentos
            </Text>
          </Group>
          <Group gap="xs" wrap="nowrap">
            <LanguageMenu />
            <Tooltip label={computedColorScheme === 'dark' ? t('Modo claro') : t('Modo escuro')}>
              <ActionIcon
                variant="subtle"
                color="gray"
                onClick={() => setColorScheme(computedColorScheme === 'dark' ? 'light' : 'dark')}
                aria-label={t('Cambiar esquema de cor')}
              >
                {computedColorScheme === 'dark' ? <IconSun size={20} /> : <IconMoon size={20} />}
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
                  label={t(item.label)}
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
                        label={t(child.label)}
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
                label={t(item.label)}
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
