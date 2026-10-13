import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Route, Routes } from 'react-router'
import { describe, expect, it } from 'vitest'

import { useMeta } from '@/api/meta'
import { AppLayout } from '@/components/layout/AppLayout'
import { LANGUAGE_STORAGE_KEY } from '@/i18n'
import { installFetch, testMeta } from '@/test/fetchMock'
import { renderWithProviders } from '@/test/utils'

const englishMeta = {
  ...testMeta,
  assetTypes: [
    { value: 'accion', label: 'Stock' },
    { value: 'indice', label: 'Index' },
    { value: 'copy_trading', label: 'Copy trading' },
    { value: 'fondo', label: 'Fund' },
  ],
}

function AssetTypeLabels() {
  const meta = useMeta()
  return (
    <ul>
      {meta.data?.assetTypes.map((type) => (
        <li key={type.value}>{type.label}</li>
      ))}
    </ul>
  )
}

describe('cambio de idioma', () => {
  it('traduce a interface e volve pedir ao servidor os datos no novo idioma', async () => {
    const calls = installFetch({
      '/api/meta': (init: RequestInit) =>
        new Headers(init.headers).get('Accept-Language') === 'en' ? englishMeta : testMeta,
    })
    renderWithProviders(
      <Routes>
        <Route element={<AppLayout />}>
          <Route index element={<AssetTypeLabels />} />
        </Route>
      </Routes>,
    )

    expect(await screen.findByText('Acción')).toBeInTheDocument()
    expect(screen.getByText('Portofolio')).toBeInTheDocument()

    const user = userEvent.setup()
    await user.click(screen.getByRole('button', { name: 'Idioma' }))
    await user.click(await screen.findByRole('menuitem', { name: 'English' }))

    expect(await screen.findByText('Stock')).toBeInTheDocument()
    expect(screen.getByText('Portfolio')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Language' })).toHaveTextContent('EN')
    expect(screen.queryByText('Acción')).not.toBeInTheDocument()
    expect(localStorage.getItem(LANGUAGE_STORAGE_KEY)).toBe('en')
    expect(document.documentElement.lang).toBe('en')
    await waitFor(() => {
      const languages = calls.filter((call) => call.url === '/api/meta').map((call) => new Headers(call.init.headers).get('Accept-Language'))
      expect(languages).toEqual(['gl', 'en'])
    })
  })
})
