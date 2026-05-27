import { computed, type MaybeRefOrGetter, toValue } from 'vue'
import type { BranchNode, BranchOverviewEntry } from '@/api/client'
import {
  buildForkTimelineLayout,
  type ForkTimelineLayout,
  type LayoutMode,
} from '@/utils/forkTimelineLayout'

export interface UseForkTimelineLayoutOptions {
  branches: MaybeRefOrGetter<BranchOverviewEntry[]>
  branchRoots: MaybeRefOrGetter<BranchNode[]>
  activeVersionId: MaybeRefOrGetter<string>
  displayActiveBranchId?: MaybeRefOrGetter<string | undefined>
  mode: MaybeRefOrGetter<LayoutMode>
  forkSequenceOverrides?: MaybeRefOrGetter<Record<string, number> | undefined>
}

export function useForkTimelineLayout(options: UseForkTimelineLayoutOptions) {
  const layout = computed<ForkTimelineLayout>(() =>
    buildForkTimelineLayout({
      branches: toValue(options.branches),
      branchRoots: toValue(options.branchRoots),
      activeVersionId: toValue(options.activeVersionId),
      displayActiveBranchId: toValue(options.displayActiveBranchId),
      mode: toValue(options.mode),
      forkSequenceOverrides: toValue(options.forkSequenceOverrides),
    })
  )

  return { layout }
}
