<script setup lang="ts">
import Search from './Search.vue';
import TableContent from './TableContent.vue';
// import Pagination from './Pagination.vue';
import type { TableProps } from '@/types';

const props = defineProps<TableProps>();

const emit = defineEmits<{
    (e: 'search', form: Record<string, any>): void
    (e: 'refresh'): void
}>()

const handleSearch = (params: Record<string, any>) => {
  emit('search', params)
}

const handleRefresh = () => {
  emit('refresh')
}

</script>

<template>
  <div class="flex flex-col gap-4">
    <!-- 标题栏 -->
    <div class="flex items-center justify-between" v-if="props.title">
      <h2 class="text-lg font-bold text-base-content">{{ props.title }}</h2>
      <button class="btn btn-link" @click="handleRefresh">
        <Icon icon="mdi:refresh" class="w-5 h-5" />
        刷新
      </button>
    </div>

    <!-- 搜索栏 -->
    <Search v-if="searchItems" :items="searchItems" @search="handleSearch" @reset="handleRefresh" />

    <!-- 表格内容 -->
    <div class="bg-base-100 rounded-lg shadow-sm border border-base-200 overflow-hidden">
      <TableContent :columns="columns" :data="data" :is-loading="isLoading" />
      
      <!-- 分页器 -->
      <div class="border-t border-base-200">
        <!-- <Pagination v-if="total && total > 0" :current-page="currentPage" :page-size="pageSize" :total="total" /> -->
      </div>
    </div>
  </div>
</template>