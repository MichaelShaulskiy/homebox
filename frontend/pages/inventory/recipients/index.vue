<script setup lang="ts">
  import type { InventoryRecipientCreate } from "~/lib/api/types/data-contracts";
  import BaseContainer from "~/components/Base/Container.vue";
  import BaseSectionHeader from "~/components/Base/SectionHeader.vue";
  import { Card } from "~/components/ui/card";
  import { Input } from "~/components/ui/input";
  import { Button } from "~/components/ui/button";
  import { Switch } from "~/components/ui/switch";

  definePageMeta({ middleware: ["auth"] });
  useHead({ title: "HomeBox | Mitarbeitende" });

  const api = useUserApi();
  const search = ref("");
  const showForm = ref(false);
  const form = ref<InventoryRecipientCreate>({
    firstName: "",
    lastName: "",
    email: "",
    department: "",
    isActive: true,
  });

  const { data: recipients, refresh } = await useAsyncData("inventory-recipients", () =>
    api.inventory.recipients().then(result => result.data || [])
  );

  const filtered = computed(() =>
    (recipients.value || []).filter(recipient =>
      `${recipient.firstName} ${recipient.lastName} ${recipient.email || ""} ${recipient.department}`
        .toLowerCase()
        .includes(search.value.toLowerCase())
    )
  );

  async function saveRecipient() {
    const response = await api.inventory.createRecipient(form.value);
    if (response.error) {
      return;
    }
    showForm.value = false;
    form.value = { firstName: "", lastName: "", email: "", department: "", isActive: true };
    await refresh();
  }
</script>

<template>
  <BaseContainer class="flex flex-col gap-4">
    <BaseSectionHeader>{{ $t("inventory.recipients") }}</BaseSectionHeader>

    <div class="flex gap-2">
      <Input v-model="search" :placeholder="$t('global.search')" />
      <Button @click="showForm = !showForm">{{ $t("inventory.add_recipient") }}</Button>
    </div>

    <Card v-if="showForm" class="p-4">
      <form class="grid gap-3 md:grid-cols-2" @submit.prevent="saveRecipient">
        <Input v-model="form.firstName" :placeholder="$t('inventory.first_name')" required />
        <Input v-model="form.lastName" :placeholder="$t('inventory.last_name')" required />
        <Input v-model="form.email" type="email" :placeholder="$t('inventory.email')" />
        <Input v-model="form.department" :placeholder="$t('inventory.department')" />
        <label class="flex items-center gap-2">
          <Switch v-model:checked="form.isActive" />
          {{ $t("inventory.active") }}
        </label>
        <Button type="submit">{{ $t("global.save") }}</Button>
      </form>
    </Card>

    <Card>
      <div class="divide-y">
        <NuxtLink
          v-for="recipient in filtered"
          :key="recipient.id"
          :to="`/inventory/recipients/${recipient.id}`"
          class="flex items-center justify-between p-4 hover:bg-accent"
        >
          <div>
            <div class="font-medium">{{ recipient.firstName }} {{ recipient.lastName }}</div>
            <div class="text-sm text-muted-foreground">
              {{ recipient.department || "-" }} <span v-if="recipient.email">· {{ recipient.email }}</span>
            </div>
          </div>
          <span :class="recipient.isActive ? 'text-green-600' : 'text-muted-foreground'">
            {{ recipient.isActive ? $t("inventory.active") : $t("inventory.inactive") }}
          </span>
        </NuxtLink>
      </div>
    </Card>
  </BaseContainer>
</template>
