import { screen } from '@testing-library/react'

import { ApiError } from '@/api/client'
import { CreateAssetForm } from '@/features/assets/forms'
import { renderWithProviders } from '@/test/utils'

describe('CreateAssetForm', () => {
  // Regresión: o efecto que copia os erros do servidor ao formulario dependía
  // do obxecto form (novo en cada render) e entraba nun bucle de renders.
  it('amosa os erros de campo do servidor sen entrar nun bucle de renders', async () => {
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {})
    const error = new ApiError('datos non válidos', 400, { name: 'o nome xa existe', month: 'mes non válido' }, new Headers())

    renderWithProviders(<CreateAssetForm error={error} onSubmit={() => {}} />)

    expect(await screen.findByText('o nome xa existe')).toBeInTheDocument()
    expect(screen.getByText('mes non válido')).toBeInTheDocument()
    const loopWarnings = consoleError.mock.calls.filter((args) => String(args[0]).includes('Maximum update depth'))
    expect(loopWarnings).toHaveLength(0)
    consoleError.mockRestore()
  })
})
