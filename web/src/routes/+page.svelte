<script>
    import {onMount} from 'svelte';
    import {writable} from 'svelte/store';
    import Modal from './Modal.svelte';

    let users = [];
    let routes = [];
    let showAddModal = false;

    let newUser = {
        id: '',
        allowedips: '',
        pre_up: '',
        post_up: '',
        pre_down: '',
        post_down: '',
        advertise_routes: '',
        accept_routes: ''
    };

    onMount(async () => {
        await fetchUsers();
        await fetchRoutes();
    });

    async function fetchUsers() {
        const response = await fetch('http://localhost:8080/api/getall', {
            method: 'POST',
        });
        const data = await response.json();
        users = data.data;
        console.log(users);
    }

    async function fetchRoutes() {
        const response = await fetch('http://localhost:8080/api/getroutes', {
            method: 'POST',
        });
        const data = await response.json();
        routes = data.data;
        console.log(routes);
    }

    async function addUser() {
        const response = await fetch('http://localhost:8080/api/adduser', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(newUser),
        });
        const data = await response.json();
        if (!response.ok) {
            alert(data.message);
        }
        await fetchUsers();
        showAddModal = false;
    }

    async function deleteUser(id) {
        const response = await fetch('http://localhost:8080/api/deluser', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({id}),
        });
        await fetchUsers();
    }

    async function copyUserConfig(id) {
        const response = await fetch('http://localhost:8080/api/getuser', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({id}),
        });
        const data = await response.json();
        const config = data.data.user_config;
        await navigator.clipboard.writeText(config);
    }
</script>

<main class="max-w-2xl mx-auto p-4 bg-gray-900 text-white rounded-lg shadow-lg">
    <h1 class="text-3xl font-bold mb-6">User List</h1>
    <div class="flex space-x-2 mb-4">
        <button on:click={fetchUsers}
                class="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 transition-all duration-200">Refresh
        </button>
        <button on:click={() => showAddModal = true}
                class="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700 transition-all duration-200">Add
            User
        </button>
    </div>

    <ul class="space-y-4">
        {#if users !== null }
            {#each users as user}
                <li class="flex items-center justify-between p-4 bg-gray-800 rounded-lg shadow">
                    <span>{user.user_id} - {user.ip}</span>
                    <div class="flex space-x-2">
                        <button on:click={() => copyUserConfig(user.user_id)}
                                class="bg-yellow-600 text-white px-4 py-2 rounded hover:bg-yellow-700 transition-all duration-200">
                            Copy Config
                        </button>
                        <button on:click={() => deleteUser(user.user_id)}
                                class="bg-red-600 text-white px-4 py-2 rounded hover:bg-red-700 transition-all duration-200">
                            Delete
                        </button>
                    </div>
                </li>
            {/each}
        {/if}
    </ul>

    {#if showAddModal}
        <Modal on:close={() => showAddModal = false}>
            <form on:submit|preventDefault={addUser} class="bg-gray-800 p-6 rounded-lg shadow-lg">
                <h2 class="text-2xl font-bold mb-4">Add User</h2>
                <div class="space-y-4">
                    <label class="block">
                        <span class="text-gray-400">ID:</span>
                        <input bind:value={newUser.id}
                               class="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white"/>
                    </label>
                    <label class="block">
                        <span class="text-gray-400">Allowed IPs:</span>
                        <input bind:value={newUser.allowedips}
                               class="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white"/>
                    </label>
                    <label class="block">
                        <span class="text-gray-400">Pre Up:</span>
                        <input bind:value={newUser.pre_up}
                               class="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white"/>
                    </label>
                    <label class="block">
                        <span class="text-gray-400">Post Up:</span>
                        <input bind:value={newUser.post_up}
                               class="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white"/>
                    </label>
                    <label class="block">
                        <span class="text-gray-400">Pre Down:</span>
                        <input bind:value={newUser.pre_down}
                               class="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white"/>
                    </label>
                    <label class="block">
                        <span class="text-gray-400">Post Down:</span>
                        <input bind:value={newUser.post_down}
                               class="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white"/>
                    </label>
                    <label class="block">
                        <span class="text-gray-400">Advertise Routes:</span>
                        <input bind:value={newUser.advertise_routes}
                               class="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white"/>
                    </label>
                    <label class="block">
                        <span class="text-gray-400">Accept Routes:</span>
                        <select bind:value={newUser.accept_routes}
                                class="mt-1 block w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white">
                            <option value="">Select a route</option>
                            {#each routes as route}
                                <option value={route}>{route}</option>
                            {/each}
                        </select>
                    </label>
                </div>
                <div class="mt-6 flex justify-end space-x-2">
                    <button type="submit"
                            class="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 transition-all duration-200">
                        Add User
                    </button>
                    <button type="button" on:click={() => showAddModal = false}
                            class="bg-gray-600 text-white px-4 py-2 rounded hover:bg-gray-700 transition-all duration-200">
                        Cancel
                    </button>
                </div>
            </form>
        </Modal>
    {/if}
</main>

<style>
    main {
        max-width: 800px;
        margin: 0 auto;
    }
</style>
