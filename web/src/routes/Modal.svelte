<script>
    import { createEventDispatcher } from 'svelte';

    const dispatch = createEventDispatcher();

    function close() {
        dispatch('close');
    }
</script>

<div class="modal" on:click|self={close}>
    <div class="modal-content relative bg-gray-800 text-white rounded-lg shadow-lg p-6">
        <button on:click={close} class="absolute top-2 right-2 text-gray-500 hover:text-gray-300 transition-all duration-200">
            &times;
        </button>
        <slot></slot>
    </div>
</div>

<style>
    .modal {
        position: fixed;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background: rgba(0, 0, 0, 0.75);
        display: flex;
        justify-content: center;
        align-items: center;
    }
    .modal-content {
        max-width: 500px;
        width: 100%;
    }
</style>
