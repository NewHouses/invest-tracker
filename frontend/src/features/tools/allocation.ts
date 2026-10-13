import type { AllocationRequest } from '@/api/tools'
import type { Asset, AssetType } from '@/api/types'

// Orde fixa dos tipos no reparto.
export const assetTypeOrder: AssetType[] = ['accion', 'indice', 'copy_trading', 'fondo']

// Activo que aínda non se creou: só serve para calcular o reparto da primeira
// vez que se inclúe.
export type NewAsset = { key: string; type: AssetType; name: string }

// Un grupo do reparto: os activos dun tipo, tanto os creados coma os novos.
export type AllocationGroup = { type: AssetType; assets: Asset[]; newAssets: NewAsset[] }

// Cada fila do reparto identifícase cunha clave: os activos creados polo seu
// id e os novos pola súa propia clave.
export function existingKey(id: number) {
  return `a${id}`
}

export function groupKeys(group: AllocationGroup) {
  return [...group.assets.map((asset) => existingKey(asset.id)), ...group.newAssets.map((item) => item.key)]
}

// buildGroups agrupa por tipo, na orde de sempre, os activos escollidos e os
// novos. Os tipos sen ningún quedan fóra.
export function buildGroups(selected: Asset[], newAssets: NewAsset[]): AllocationGroup[] {
  return assetTypeOrder
    .map((type) => ({ type, assets: selected.filter((asset) => asset.type === type), newAssets: newAssets.filter((item) => item.type === type) }))
    .filter((group) => group.assets.length + group.newAssets.length > 0)
}

// equalPercents propón o reparto de sempre (o total a partes iguais entre os
// tipos e, dentro de cada tipo, entre os seus activos) coas porcentaxes
// redondeadas a dúas decimais, que é como se amosan e se editan.
export function equalPercents(groups: AllocationGroup[]): Record<string, number> {
  const out: Record<string, number> = {}
  for (const group of groups) {
    const keys = groupKeys(group)
    for (const key of keys) out[key] = Math.round((100 / groups.length / keys.length) * 100) / 100
  }
  return out
}

export function validPercent(value: number) {
  return Number.isFinite(value) && value >= 0 && value <= 100
}

// sumsTo100 indica se n porcentaxes que suman sum suman 100, admitindo o que
// se perde ao escribir cada unha con dúas decimais (33,33 × 3 = 99,99), igual
// ca o servidor.
export function sumsTo100(sum: number, n: number) {
  return Math.abs(sum - 100) <= 0.005 * n + 1e-9
}

// buildAllocationRequest prepara a petición ao servidor. Sen porcentaxes (o
// usuario non cambiou a proposta) o servidor reparte a partes iguais exactas.
export function buildAllocationRequest(total: number, groups: AllocationGroup[], percents: Record<string, number> | null): AllocationRequest {
  return {
    total,
    selection: groups.map((group) => ({
      type: group.type,
      assetIds: group.assets.map((asset) => asset.id),
      ...(percents ? { percents: group.assets.map((asset) => percents[existingKey(asset.id)] ?? Number.NaN) } : {}),
      newAssets: group.newAssets.map((item) => (percents ? { name: item.name, percent: percents[item.key] ?? Number.NaN } : { name: item.name })),
    })),
  }
}
