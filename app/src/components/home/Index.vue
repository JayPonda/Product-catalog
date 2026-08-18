<template>
    this is a {{ curruntPage }} page
    <br />
    <br />
    <br />
    <br />
    Do laborum pariatur laboris pariatur velit dolore consequat. Non do labore officia duis. Labore nisi quis id
    voluptate pariatur. Culpa cupidatat excepteur duis cillum amet fugiat. Sit Lorem voluptate quis dolor pariatur
    cillum ea eiusmod dolore deserunt eiusmod excepteur quis. Nostrud adipisicing irure in sit enim quis proident. Ipsum
    enim pariatur occaecat ipsum ea do est excepteur anim enim quis anim.
    <br />
    <br />
    <br />
    <br />
    Laborum mollit irure sint cillum. Labore esse ad nostrud occaecat adipisicing sunt voluptate laborum incididunt
    commodo mollit sunt. Labore mollit tempor ex minim nostrud nulla eu incididunt sit quis. Occaecat eiusmod est eu
    duis quis amet sint pariatur in anim. Irure laboris eiusmod minim aliqua qui voluptate in consectetur excepteur esse
    culpa. Cillum do eiusmod nisi reprehenderit tempor culpa quis consequat.
    <br />
    <br />
    <br />
    <br />
    Do laborum pariatur laboris pariatur velit dolore consequat. Non do labore officia duis. Labore nisi quis id
    voluptate pariatur. Culpa cupidatat excepteur duis cillum amet fugiat. Sit Lorem voluptate quis dolor pariatur
    cillum ea eiusmod dolore deserunt eiusmod excepteur quis. Nostrud adipisicing irure in sit enim quis proident. Ipsum
    enim pariatur occaecat ipsum ea do est excepteur anim enim quis anim.
    <br />
    <br />
    <br />
    <br />
    Laborum mollit irure sint cillum. Labore esse ad nostrud occaecat adipisicing sunt voluptate laborum incididunt
    commodo mollit sunt. Labore mollit tempor ex minim nostrud nulla eu incididunt sit quis. Occaecat eiusmod est eu
    duis quis amet sint pariatur in anim. Irure laboris eiusmod minim aliqua qui voluptate in consectetur excepteur esse
    culpa. Cillum do eiusmod nisi reprehenderit tempor culpa quis consequat.
    <br />
    <br />
    <br />
    <br />
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useRoute } from 'vue-router';
const route = useRoute()

const curruntPage = computed(function(){
    const params = route.path.slice(1).split("/", 2)
    if(params.length == 0 || (params.length > 0 && params[0]?.trim() == '')){
        return 'Home';
    }
    const routename = params[0]?.trim().replaceAll('-', ' ').toLowerCase() ?? 'Home';
    return routename[0]?.toUpperCase() + routename.slice(1)
})

</script>