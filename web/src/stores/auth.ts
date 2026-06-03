import { defineStore } from "pinia";
import type { LoginRequest, UserResponse } from "@/types";
import { login as loginApi, getUser as getUserApi } from "@/api/user";

export const useUserStore = defineStore(
  "user",
  () => {
    const token = ref<string | null>(null);
    const user = ref<UserResponse | null>(null);
    const currentWorkspace = ref<string | null>(null);
    const currentApp = ref<string | null>(null);

    const isLogined = computed(() => !!token.value);

    const workspaceList = computed(() => user.value?.workspace_list || []);
    const appList = computed(
      () =>
        workspaceList.value.find(
          (workspace) => String(workspace.id) === String(currentWorkspace.value)
        )?.app_list
    );

    const login = async (data: LoginRequest): Promise<void> => {
      const res = await loginApi(data);
      token.value = res.token;

      const userInfoRes = await getUserApi();
      user.value = userInfoRes;

      currentWorkspace.value = userInfoRes.workspace_list[0]?.id || null;
      currentApp.value =
        userInfoRes.workspace_list[0]?.app_list[0]?.id || null;
    };

    const logout = () => {
      token.value = null;
      user.value = null;
      currentWorkspace.value = null;
      currentApp.value = null;
    };

    const changeWorkspace = (workspaceId?: string, appId?: string) => {
      if (workspaceId) {
        currentWorkspace.value = workspaceId;
      }
      if (appId) {
        currentApp.value = appId;
      }
    };

    return {
      token,
      user,
      isLogined,
      login,
      logout,
      currentWorkspace,
      currentApp,
      workspaceList,
      appList,
      changeWorkspace,
    };
  },
  {
    persist: true,
  }
);
