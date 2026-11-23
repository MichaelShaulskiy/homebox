<script setup lang="ts">
  import { useI18n } from "vue-i18n";
  import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
  import { Button } from "@/components/ui/button";
  import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
  import MdiDelete from "~icons/mdi/delete";
  import { toast } from "@/components/ui/sonner";
  import type { UserAccountCreate, UserSummary } from "~~/lib/api/types/data-contracts";
  import { UserAccountRole } from "~~/lib/api/types/data-contracts";

  definePageMeta({
    middleware: ["auth"],
  });

  const { t } = useI18n();

  useHead({ title: `HomeBox | ${t("collection.tabs.members")}` });

  const api = useUserApi();
  const auth = useAuthContext();
  const confirm = useConfirm();

  const loading = ref(true);
  const members = ref<UserSummary[]>([]);
  const error = ref<string | null>(null);
  const removing = ref<Record<string, boolean>>({});
  const supervisors = ref<UserSummary[]>([]);
  const showCreate = ref(false);
  const createForm = ref<UserAccountCreate>({
    name: "",
    email: "",
    password: "",
    accountRole: UserAccountRole.AccountRoleUser,
    supervisorIds: [],
  });

  const currentUserId = computed(() => auth.user?.id ?? "");
  const canManageAccounts = computed(
    () =>
      auth.user?.accountRole === UserAccountRole.AccountRoleAdmin ||
      auth.user?.accountRole === UserAccountRole.AccountRoleSuperadmin ||
      auth.user?.isSuperuser
  );

  const isLastMember = computed(() => members.value.length <= 1);

  const loadMembers = async () => {
    loading.value = true;
    error.value = null;

    try {
      const res = canManageAccounts.value ? await api.user.getAccountMembers() : await api.group.getMembers();
      if (res.error) {
        const msg = t("errors.api_failure") + String(res.error);
        error.value = msg;
        members.value = [];
        toast.error(msg);
      } else {
        members.value = res.data ?? [];
      }
      if (canManageAccounts.value) {
        const supervisorRes = await api.user.getAccountSupervisors();
        supervisors.value = supervisorRes.data ?? [];
      }
    } catch (e) {
      const msg = (e as Error).message ?? String(e);
      error.value = msg;
      members.value = [];
      toast.error(msg);
    } finally {
      loading.value = false;
    }
  };

  const handleRemove = async (user: UserSummary) => {
    if (!user?.id) return;

    if (isLastMember.value && user.id === currentUserId.value) {
      toast.error(t("collection.members.cannot_remove_last"));
      return;
    }

    const result = await confirm.open(t("collection.members.remove_confirm"));
    if (result.isCanceled) {
      return;
    }

    removing.value = { ...removing.value, [user.id]: true };

    try {
      const res = await api.group.removeMember(user.id);
      if (res.error) {
        const msg = t("errors.api_failure") + String(res.error);
        toast.error(msg);
      } else {
        members.value = members.value.filter(m => m.id !== user.id);
        toast.success(t("collection.members.removed"));
      }
    } catch (e) {
      const msg = (e as Error).message ?? String(e);
      toast.error(msg);
    } finally {
      removing.value = { ...removing.value, [user.id]: false };
    }
  };

  const createAccount = async () => {
    if (!canManageAccounts.value) return;
    const res = await api.user.createAccount(createForm.value);
    if (res.error) {
      toast.error(t("errors.api_failure"));
      return;
    }
    showCreate.value = false;
    createForm.value = {
      name: "",
      email: "",
      password: "",
      accountRole: UserAccountRole.AccountRoleUser,
      supervisorIds: [],
    };
    await loadMembers();
  };

  const setRole = async (user: UserSummary, role: string) => {
    const res = await api.user.setAccountRole(user.id, {
      accountRole: role as UserAccountRole,
    });
    if (res.error) {
      toast.error(t("errors.api_failure"));
      return;
    }
    await loadMembers();
  };

  const setRoleFromEvent = (user: UserSummary, event: Event) => {
    const target = event.target as HTMLSelectElement;
    void setRole(user, target.value);
  };

  const setSupervisors = async (user: UserSummary, ids: string[]) => {
    const res = await api.user.setAccountSupervisors(user.id, {
      supervisorIds: ids,
    });
    if (res.error) {
      toast.error(t("errors.api_failure"));
      return;
    }
    await loadMembers();
  };

  const toggleSupervisor = (user: UserSummary, supervisorId: string, checked: boolean) => {
    const ids = checked ? [...user.supervisorIds, supervisorId] : user.supervisorIds.filter(id => id !== supervisorId);
    void setSupervisors(user, ids);
  };

  const toggleSupervisorFromEvent = (user: UserSummary, supervisorId: string, event: Event) => {
    const target = event.target as HTMLInputElement;
    toggleSupervisor(user, supervisorId, target.checked);
  };

  onMounted(() => {
    loadMembers();
  });
</script>

<template>
  <div class="space-y-4">
    <div v-if="canManageAccounts" class="rounded-md border bg-card p-4">
      <div class="mb-3 flex items-center justify-between">
        <span class="font-medium">{{ $t("accounts.title") }}</span>
        <Button size="sm" @click="showCreate = !showCreate">{{ $t("accounts.create") }}</Button>
      </div>
      <form v-if="showCreate" class="grid gap-3 md:grid-cols-2" @submit.prevent="createAccount">
        <input
          v-model="createForm.name"
          class="h-10 rounded-md border bg-background px-3"
          :placeholder="$t('accounts.name')"
          required
        />
        <input
          v-model="createForm.email"
          type="email"
          class="h-10 rounded-md border bg-background px-3"
          :placeholder="$t('accounts.email')"
          required
        />
        <input
          v-model="createForm.password"
          type="password"
          class="h-10 rounded-md border bg-background px-3"
          :placeholder="$t('accounts.password')"
          required
        />
        <select v-model="createForm.accountRole" class="h-10 rounded-md border bg-background px-3">
          <option :value="UserAccountRole.AccountRoleAdmin">{{ $t("accounts.admin") }}</option>
          <option :value="UserAccountRole.AccountRoleUser">{{ $t("accounts.user") }}</option>
        </select>
        <div class="flex flex-wrap gap-3 md:col-span-2">
          <label v-for="supervisor in supervisors" :key="supervisor.id" class="flex items-center gap-2 text-sm">
            <input v-model="createForm.supervisorIds" type="checkbox" :value="supervisor.id" />
            {{ supervisor.name }}
          </label>
        </div>
        <Button type="submit">{{ $t("accounts.create") }}</Button>
      </form>
    </div>

    <div v-if="loading" class="rounded-md border bg-card p-4 text-sm text-muted-foreground">
      {{ $t("global.loading") }}
    </div>

    <div v-else>
      <div v-if="!members.length" class="rounded-md border bg-card p-4 text-sm text-muted-foreground">
        {{ $t("collection.members.empty") }}
      </div>

      <div v-else class="scroll-bg overflow-x-auto rounded-md border bg-card">
        <Table class="min-w-[480px]">
          <TableHeader>
            <TableRow>
              <TableHead>{{ $t("collection.members.name") }}</TableHead>
              <TableHead>{{ $t("collection.members.email") }}</TableHead>
              <TableHead>{{ $t("accounts.role") }}</TableHead>
              <TableHead>{{ $t("accounts.supervisors") }}</TableHead>
              <TableHead class="w-32 text-right"></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="user in members" :key="user.id">
              <TableCell>{{ user.name }}</TableCell>
              <TableCell>{{ user.email }}</TableCell>
              <TableCell>
                <select
                  v-if="canManageAccounts && user.accountRole !== UserAccountRole.AccountRoleSuperadmin"
                  :value="user.accountRole"
                  class="h-9 rounded-md border bg-background px-2"
                  @change="setRoleFromEvent(user, $event)"
                >
                  <option :value="UserAccountRole.AccountRoleAdmin">{{ $t("accounts.admin") }}</option>
                  <option :value="UserAccountRole.AccountRoleUser">{{ $t("accounts.user") }}</option>
                </select>
                <span v-else>{{ $t(`accounts.${user.accountRole}`) }}</span>
              </TableCell>
              <TableCell>
                <div
                  v-if="canManageAccounts && user.accountRole !== UserAccountRole.AccountRoleSuperadmin"
                  class="flex flex-wrap gap-2"
                >
                  <label v-for="supervisor in supervisors" :key="supervisor.id" class="flex items-center gap-1 text-xs">
                    <input
                      type="checkbox"
                      :checked="user.supervisorIds.includes(supervisor.id)"
                      @change="toggleSupervisorFromEvent(user, supervisor.id, $event)"
                    />
                    {{ supervisor.name }}
                  </label>
                </div>
                <span v-else>-</span>
              </TableCell>
              <TableCell>
                <div class="ml-auto">
                  <TooltipProvider :delay-duration="0">
                    <Tooltip>
                      <TooltipTrigger as-child>
                        <Button
                          v-if="canManageAccounts"
                          variant="destructive"
                          size="icon"
                          :aria-label="$t('global.delete')"
                          :disabled="removing[user.id] || (isLastMember && user.id === currentUserId)"
                          @click="handleRemove(user)"
                        >
                          <MdiDelete class="size-4" />
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent>
                        {{ $t("global.delete") }}
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                </div>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>
    </div>
  </div>
</template>
