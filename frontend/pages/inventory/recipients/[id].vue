<script setup lang="ts">
  import BaseContainer from "~/components/Base/Container.vue";
  import BaseSectionHeader from "~/components/Base/SectionHeader.vue";
  import { Card } from "~/components/ui/card";
  import { Button } from "~/components/ui/button";

  definePageMeta({ middleware: ["auth"] });

  const route = useRoute();
  const api = useUserApi();
  const recipientId = computed(() => route.params.id as string);

  const { data: recipient, refresh } = await useAsyncData(`recipient-${recipientId.value}`, () =>
    api.inventory.recipient(recipientId.value).then(result => result.data)
  );
  const { data: assignments, refresh: refreshAssignments } = await useAsyncData(
    `recipient-assignment-${recipientId.value}`,
    () => api.inventory.assignments({ recipientId: recipientId.value }).then(result => result.data || [])
  );
  const { data: transactions, refresh: refreshTransactions } = await useAsyncData(
    `recipient-history-${recipientId.value}`,
    () => api.inventory.recipientTransactions(recipientId.value).then(result => result.data || [])
  );

  async function deactivate() {
    if (!recipient.value) return;
    const response = await api.inventory.deactivateRecipient(recipient.value.id);
    if (!response.error) await refresh();
  }

  async function returnItem(assignmentId: string) {
    await api.inventory.returnAssignment(assignmentId, {
      assignmentId,
      quantity: 1,
      status: "OK",
      reasonOrNote: "",
    });
    await refreshAssignments();
    await refreshTransactions();
  }
</script>

<template>
  <BaseContainer v-if="recipient" class="flex flex-col gap-4">
    <BaseSectionHeader>{{ recipient.firstName }} {{ recipient.lastName }}</BaseSectionHeader>
    <Card class="p-4">
      <p>{{ recipient.department || "-" }} · {{ recipient.email || "-" }}</p>
      <Button v-if="recipient.isActive" variant="destructive" class="mt-3" @click="deactivate">
        {{ $t("inventory.deactivate") }}
      </Button>
    </Card>

    <Card class="p-4">
      <h2 class="mb-3 font-semibold">{{ $t("inventory.current_assignments") }}</h2>
      <div v-if="!assignments?.length" class="text-muted-foreground">{{ $t("inventory.none") }}</div>
      <div
        v-for="assignment in assignments"
        :key="assignment.id"
        class="flex items-center justify-between border-t py-3"
      >
        <div>
          <NuxtLink class="font-medium" :to="`/item/${assignment.itemId}`">{{ assignment.itemName }}</NuxtLink>
          <div class="text-sm text-muted-foreground">
            {{ assignment.quantity }}x · {{ assignment.assetTag || assignment.serialNumber || "-" }}
          </div>
        </div>
        <Button size="sm" @click="returnItem(assignment.id)">{{ $t("inventory.return_ok") }}</Button>
      </div>
    </Card>

    <Card class="p-4">
      <h2 class="mb-3 font-semibold">{{ $t("inventory.history") }}</h2>
      <div v-for="row in transactions" :key="row.id" class="border-t py-3">
        <div class="font-medium">{{ row.actionType }} · {{ row.itemName }} ({{ row.quantity }})</div>
        <div class="text-sm text-muted-foreground">
          {{ row.timestamp }} · {{ row.performedByAdminName }}
          <span v-if="row.reasonOrNote"> · {{ row.reasonOrNote }}</span>
        </div>
      </div>
    </Card>
  </BaseContainer>
</template>
