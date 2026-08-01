<template>
  <section>
    <div class="page-heading"><h1>Usuarios</h1><p>Crea cuentas y administra los datos, roles y acceso de cada persona.</p></div>
    <div class="toolbar">
      <div class="form-field"><label for="search">Buscar</label><InputText id="search" v-model.trim="search" placeholder="Correo, nombre o apellido" @keyup.enter="applySearch" /></div>
      <div class="form-field"><label for="role-filter">Rol</label><Select input-id="role-filter" v-model="roleFilter" :options="roleOptions" option-label="label" option-value="value" placeholder="Todos" show-clear @change="applySearch" /></div>
      <div class="form-field"><label for="status-filter">Estado</label><Select input-id="status-filter" v-model="statusFilter" :options="statusOptions" option-label="label" option-value="value" placeholder="Todos" show-clear @change="applySearch" /></div>
      <Button label="Crear usuario" icon="pi pi-user-plus" @click="openCreate" />
      <Button label="Buscar" icon="pi pi-search" :loading="loading" @click="applySearch" />
    </div>
    <Message v-if="errorMessage" severity="error" :closable="false">{{ errorMessage }}</Message>
    <UiLoadingState v-if="loading" />
    <UiEmptyState v-else-if="users.length === 0" title="No hay usuarios" description="Ajusta los filtros o crea una cuenta nueva." />
    <template v-else>
      <div class="table-wrap">
        <DataTable :value="users" data-key="id" striped-rows>
          <Column field="email" header="Correo" />
          <Column header="Nombre"><template #body="{ data }">{{ data.first_name }} {{ data.last_name }}</template></Column>
          <Column header="Rol"><template #body="{ data }"><Tag :value="roleLabel(data.role)" :severity="data.role === 'admin' ? 'info' : 'secondary'" /></template></Column>
          <Column header="Estado"><template #body="{ data }"><Tag :value="statusLabel(data.status)" :severity="data.status === 'active' ? 'success' : 'danger'" /></template></Column>
          <Column header="Acciones"><template #body="{ data }"><Button label="Ver" icon="pi pi-eye" size="small" outlined @click="openUser(data.id)" /></template></Column>
        </DataTable>
      </div>
      <Paginator :rows="pagination.limit" :total-records="pagination.total" :first="(pagination.page - 1) * pagination.limit" @page="changePage" />
    </template>

    <Dialog v-model:visible="detailVisible" modal header="Detalle de usuario" :style="{ width: 'min(34rem, calc(100vw - 2rem))' }">
      <UiLoadingState v-if="detailLoading" />
      <template v-else-if="selected">
        <Message v-if="detailError" severity="error" :closable="false">{{ detailError }}</Message>
        <dl class="detail-grid">
          <div class="detail-item"><dt>Nombre</dt><dd>{{ selected.first_name }} {{ selected.last_name }}</dd></div>
          <div class="detail-item"><dt>Correo</dt><dd>{{ selected.email }}</dd></div>
          <div class="detail-item"><dt>Creado</dt><dd>{{ formatDate(selected.created_at) }}</dd></div>
          <div class="detail-item"><dt>Último ingreso</dt><dd>{{ formatDate(selected.last_login_at) }}</dd></div>
        </dl>
        <div class="stack" style="margin-top: var(--space-4)">
          <div class="form-field"><label for="edit-name">Nombre</label><InputText id="edit-name" v-model.trim="selected.first_name" :disabled="isSelf || saving" maxlength="100" /></div>
          <div class="form-field"><label for="edit-lastname">Apellido</label><InputText id="edit-lastname" v-model.trim="selected.last_name" :disabled="isSelf || saving" maxlength="100" /></div>
          <div class="form-field"><label for="edit-phone">Teléfono</label><InputText id="edit-phone" v-model.trim="selected.phone" :disabled="isSelf || saving" maxlength="30" /></div>
          <div class="form-field"><label for="edit-document">Documento</label><InputText id="edit-document" v-model.trim="selected.document_number" :disabled="isSelf || saving" maxlength="40" /></div>
          <div class="form-field"><label for="edit-nationality">Nacionalidad</label><InputText id="edit-nationality" v-model.trim="selected.nationality" :disabled="isSelf || saving" maxlength="80" /></div>
        </div>
        <div class="detail-grid" style="margin-top: var(--space-6)">
          <div class="form-field"><label for="role">Rol</label><Select id="role" v-model="selected.role" :options="roleOptions" option-label="label" option-value="value" :disabled="isSelf || saving" /></div>
          <div class="form-field"><label for="status">Estado</label><Select id="status" v-model="selected.status" :options="statusOptions" option-label="label" option-value="value" :disabled="isSelf || saving" /></div>
        </div>
        <Message v-if="isSelf" severity="info" :closable="false">Por seguridad no puedes cambiar tu propio rol o estado desde esta pantalla.</Message>
        <div class="cluster" style="justify-content: flex-end; margin-top: var(--space-6)">
          <Button label="Cerrar" severity="secondary" outlined @click="detailVisible = false" />
          <Button label="Guardar" icon="pi pi-save" :disabled="isSelf || saving" :loading="saving" @click="confirmSave" />
        </div>
      </template>
    </Dialog>
    <Dialog v-model:visible="createVisible" modal header="Crear usuario" :style="{ width: 'min(34rem, calc(100vw - 2rem))' }" :closable="!creating">
      <Message v-if="createError" severity="error" :closable="false">{{ createError }}</Message>
      <form class="stack" @submit.prevent="createUser">
        <div class="form-field"><label for="create-name">Nombre</label><InputText id="create-name" v-model.trim="creation.first_name" required minlength="2" maxlength="100" /></div>
        <div class="form-field"><label for="create-lastname">Apellido</label><InputText id="create-lastname" v-model.trim="creation.last_name" required minlength="2" maxlength="100" /></div>
        <div class="form-field"><label for="create-email">Correo electrónico</label><InputText id="create-email" v-model.trim="creation.email" type="email" required maxlength="320" /></div>
        <div class="form-field"><label for="create-role">Rol</label><Select input-id="create-role" v-model="creation.role" :options="roleOptions" option-label="label" option-value="value" /></div>
        <div class="form-field"><label for="create-password">Contraseña inicial</label><Password input-id="create-password" v-model="creation.password" autocomplete="new-password" toggle-mask /><small class="form-help">Al menos 12 caracteres. Entrega la contraseña al usuario por un medio privado.</small></div>
        <Button type="submit" label="Crear usuario" icon="pi pi-user-plus" :loading="creating" />
      </form>
    </Dialog>
  </section>
</template>

<script setup lang="ts">
import Button from 'primevue/button'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import Paginator from 'primevue/paginator'
import Password from 'primevue/password'
import { roleOptions, roleLabel } from '~/utils/roles'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import { useConfirm } from 'primevue/useconfirm'
import { useToast } from 'primevue/usetoast'
import type { Role, User, UserStatus } from '~/types/auth'
import { apiError } from '~/utils/api-error'

interface UserPage { users: User[], pagination: { page: number, limit: number, total: number, total_pages: number } }
definePageMeta({ layout: 'admin', middleware: 'admin' })
useHead({ title: 'Usuarios' })
const auth = useAuthStore()
const confirm = useConfirm()
const toast = useToast()
const loading = ref(true)
const detailLoading = ref(false)
const saving = ref(false)
const errorMessage = ref('')
const search = ref('')
const users = ref<User[]>([])
const pagination = reactive({ page: 1, limit: 20, total: 0, total_pages: 0 })
const detailVisible = ref(false)
const selected = ref<User | null>(null)
const roleFilter = ref<Role | null>(null)
const statusFilter = ref<UserStatus | null>(null)
const detailError = ref('')
const createVisible = ref(false)
const creating = ref(false)
const createError = ref('')
const creation = reactive({ first_name: '', last_name: '', email: '', password: '', role: 'turista' as Role })
const isSelf = computed(() => selected.value?.id === auth.user?.id)
const statusOptions = [{ label: 'Activo', value: 'active' }, { label: 'Inactivo', value: 'inactive' }]

const load = async () => {
  loading.value = true
  errorMessage.value = ''
  try {
    const query = new URLSearchParams({ page: String(pagination.page), limit: String(pagination.limit) })
    if (search.value) query.set('search', search.value)
    if (roleFilter.value) query.set('role', roleFilter.value)
    if (statusFilter.value) query.set('status', statusFilter.value)
    const response = await auth.request<UserPage>(`/admin/users?${query}`)
    users.value = response.data?.users || []
    Object.assign(pagination, response.data?.pagination)
  } catch (error) { errorMessage.value = apiError(error).message }
  finally { loading.value = false }
}
const applySearch = () => { pagination.page = 1; void load() }
const changePage = (event: { page: number }) => { pagination.page = event.page + 1; void load() }
const openUser = async (id: number) => {
  selected.value = null
  detailError.value = ''
  detailVisible.value = true
  detailLoading.value = true
  try {
    const response = await auth.request<User>(`/admin/users/${id}`)
    selected.value = response.data || null
  } catch (error) { errorMessage.value = apiError(error).message; detailVisible.value = false }
  finally { detailLoading.value = false }
}
const confirmSave = () => confirm.require({
  message: 'Los cambios de rol o desactivación pueden cerrar sesiones existentes. ¿Continuar?',
  header: 'Confirmar cambios', icon: 'pi pi-exclamation-triangle', acceptLabel: 'Guardar', rejectLabel: 'Cancelar',
  accept: save
})
const save = async () => {
  if (!selected.value || saving.value) return
  detailError.value = ''
  saving.value = true
  try {
    const u = selected.value
    await auth.request(`/admin/users/${u.id}`, { method: 'PATCH', body: { first_name: u.first_name, last_name: u.last_name, phone: u.phone, document_number: u.document_number, nationality: u.nationality, role: u.role, status: u.status } })
    toast.add({ severity: 'success', summary: 'Usuario actualizado', life: 3500 })
    detailVisible.value = false
    await load()
  } catch (error) { detailError.value = apiError(error).message }
  finally { saving.value = false }
}
const formatDate = (value: string | null) => value ? new Intl.DateTimeFormat('es-BO', { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value)) : 'Sin registro'
const statusLabel = (status: UserStatus) => status === 'active' ? 'Activo' : 'Inactivo'
const openCreate = () => {
  Object.assign(creation, { first_name: '', last_name: '', email: '', password: '', role: 'turista' })
  createError.value = ''; createVisible.value = true
}
const createUser = async () => {
  if (creating.value) return
  creating.value = true; createError.value = ''
  try {
    await auth.request('/admin/users', { method: 'POST', body: creation })
    creation.password = ''
    createVisible.value = false
    toast.add({ severity: 'success', summary: 'Usuario creado', life: 3500 })
    pagination.page = 1
    await load()
  } catch (error) { createError.value = apiError(error).message }
  finally { creating.value = false }
}
onMounted(load)
</script>
