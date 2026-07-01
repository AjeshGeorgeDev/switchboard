export interface CatalogSection {
  id: string
  name: string
  sort_order: number
}

export function appSectionId(app: { section_id?: unknown }): string | null {
  const value = app.section_id
  if (!value) return null
  if (typeof value === 'string') return value
  if (typeof value === 'object' && value !== null) {
    const record = value as { Valid?: boolean; Bytes?: string }
    if (record.Valid === false) return null
    if (typeof record.Bytes === 'string' && record.Bytes) return record.Bytes
  }
  return null
}

export interface CatalogSectionGroup {
  id: string | null
  name: string
  sort_order: number
  apps: any[]
}

export function groupAppsBySection(apps: any[], sections: CatalogSection[]): CatalogSectionGroup[] {
  const buckets = new Map<string | null, any[]>()

  for (const app of apps) {
    const sectionId = appSectionId(app)
    if (!buckets.has(sectionId)) buckets.set(sectionId, [])
    buckets.get(sectionId)!.push(app)
  }

  const groups: CatalogSectionGroup[] = []
  const orderedSections = [...sections].sort(
    (a, b) => a.sort_order - b.sort_order || a.name.localeCompare(b.name),
  )

  for (const section of orderedSections) {
    const sectionApps = buckets.get(section.id)
    if (sectionApps?.length) {
      groups.push({
        id: section.id,
        name: section.name,
        sort_order: section.sort_order,
        apps: sectionApps,
      })
      buckets.delete(section.id)
    }
  }

  const uncategorized = buckets.get(null) ?? []
  if (uncategorized.length) {
    groups.push({
      id: null,
      name: 'Other',
      sort_order: 9999,
      apps: uncategorized,
    })
  }

  return groups
}

export function shouldShowSectionHeaders(groups: CatalogSectionGroup[]): boolean {
  return groups.length > 1 || (groups.length === 1 && groups[0].id !== null)
}
