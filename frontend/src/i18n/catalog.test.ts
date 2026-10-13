import { describe, expect, it } from 'vitest'

import { en, enParts } from '@/i18n/en'

// Todo o código da web agás os tests e as propias traducións.
const sources = import.meta.glob<string>(
  ['/src/**/*.{ts,tsx}', '!/src/**/*.test.{ts,tsx}', '!/src/test/**', '!/src/i18n/en.ts', '!/src/i18n/en/**'],
  { query: '?raw', import: 'default', eager: true },
)

// Chamadas t('...') e tKey('...') cun texto literal entre comiñas.
const callRe = /\b(?:t|tKey)\(\s*(['"])((?:\\.|(?!\1)[^\\])*)\1/g
// As claves teñen que ser literais: os valores van como parámetros {nome}.
const templateRe = /\b(?:t|tKey)\(\s*`/g

function unescapeJs(body: string) {
  return body.replace(/\\(u\{[0-9a-fA-F]+\}|u[0-9a-fA-F]{4}|x[0-9a-fA-F]{2}|[\s\S])/g, (_, seq: string) => {
    if (seq.startsWith('u{')) return String.fromCodePoint(parseInt(seq.slice(2, -1), 16))
    if (seq.length > 1) return String.fromCharCode(parseInt(seq.slice(1), 16))
    return ({ n: '\n', t: '\t', r: '\r' } as Record<string, string>)[seq] ?? seq
  })
}

function usedKeys() {
  const keys = new Map<string, string[]>()
  for (const [file, code] of Object.entries(sources)) {
    for (const match of code.matchAll(callRe)) {
      const key = unescapeJs(match[2] ?? '')
      keys.set(key, [...(keys.get(key) ?? []), file])
    }
  }
  return keys
}

function placeholders(text: string) {
  return [...text.matchAll(/\{(\w+)\}/g)].map((m) => m[1]).sort().join(',')
}

describe('catálogo de traducións', () => {
  const used = usedKeys()

  it('atopa os textos da web', () => {
    expect(Object.keys(sources).length).toBeGreaterThan(20)
    expect(used.size).toBeGreaterThan(0)
  })

  it('ten tradución inglesa para cada texto', () => {
    const missing = [...used.entries()]
      .filter(([key]) => !en[key]?.trim())
      .map(([key, files]) => `${key}  (${[...new Set(files)].join(', ')})`)
    expect(missing).toEqual([])
  })

  it('non garda traducións que xa non se usan', () => {
    expect(Object.keys(en).filter((key) => !used.has(key))).toEqual([])
  })

  it('usa os mesmos parámetros nos dous idiomas', () => {
    const mismatched = Object.entries(en)
      .filter(([gl, english]) => placeholders(gl) !== placeholders(english))
      .map(([gl, english]) => `${gl} → ${english}`)
    expect(mismatched).toEqual([])
  })

  it('traduce igual un texto repetido en varias áreas', () => {
    const seen = new Map<string, { part: string; english: string }>()
    const conflicts: string[] = []
    for (const [part, dictionary] of Object.entries(enParts)) {
      for (const [gl, english] of Object.entries(dictionary)) {
        const previous = seen.get(gl)
        if (previous && previous.english !== english) {
          conflicts.push(`${gl}: ${previous.part}="${previous.english}" / ${part}="${english}"`)
        }
        seen.set(gl, { part, english })
      }
    }
    expect(conflicts).toEqual([])
  })

  it('só traduce textos literais', () => {
    const dynamic = Object.entries(sources).flatMap(([file, code]) => [...code.matchAll(templateRe)].map(() => file))
    expect(dynamic).toEqual([])
  })
})
