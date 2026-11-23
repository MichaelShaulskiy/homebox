<script setup lang="ts">
  import BaseContainer from "~/components/Base/Container.vue";
  import BaseSectionHeader from "~/components/Base/SectionHeader.vue";
  import { Card } from "~/components/ui/card";
  import { Input } from "~/components/ui/input";
  import { Button } from "~/components/ui/button";

  definePageMeta({ middleware: ["auth"] });
  useHead({ title: "HomeBox | Bewegungslog" });

  const api = useUserApi();
  const search = ref("");
  const actionType = ref("");
  const { data: rows, refresh } = await useAsyncData("inventory-history", () =>
    api.inventory.transactions().then(result => result.data || [])
  );

  const filtered = computed(() =>
    (rows.value || []).filter(row => {
      const matchesSearch =
        !search.value ||
        `${row.itemName} ${row.targetUserName} ${row.performedByAdminName} ${row.reasonOrNote}`
          .toLowerCase()
          .includes(search.value.toLowerCase());
      return matchesSearch && (!actionType.value || row.actionType === actionType.value);
    })
  );
</script>

<template>
  <BaseContainer class="flex flex-col gap-4">
    <BaseSectionHeader>{{ $t("inventory.history") }}</BaseSectionHeader>
    <div class="flex flex-wrap gap-2">
      <Input v-model="search" :placeholder="$t('global.search')" />
      <select v-model="actionType" class="h-10 rounded-md border bg-background px-3">
        <option value="">Alle Aktionen</option>
        <option value="ISSUED">Herausgabe</option>
        <option value="RETURNED_OK">Rückgabe OK</option>
        <option value="RETURNED_DEFECTIVE">Rückgabe defekt</option>
        <option value="WRITTEN_OFF">Ausbuchung</option>
        <option value="STOCK_ADDED">Zugang</option>
      </select>
      <a :href="api.authURL(api.inventory.exportURL())" class="inline-flex h-10 items-center rounded-md border px-4">
        CSV
      </a>
      <Button variant="outline" @click="refresh">{{ $t("global.refresh") }}</Button>
    </div>
    <Card>
      <div v-for="row in filtered" :key="row.id" class="border-b p-4">
        <div class="font-medium">{{ row.actionType }} · {{ row.itemName }} ({{ row.quantity }})</div>
        <div class="text-sm text-muted-foreground">
          {{ row.timestamp }}
          <span v-if="row.targetUserName"> · {{ row.targetUserName }}</span>
          · {{ row.performedByAdminName }}
          <span v-if="row.reasonOrNote"> · {{ row.reasonOrNote }}</span>
        </div>
      </div>
    </Card>
  </BaseContainer>
</template>
