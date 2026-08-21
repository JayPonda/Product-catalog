
function getBaseUrl(prefix, url, queryparams) {
    const finalUrl = new URL(import.meta.env.VITE_BACKEND_URL + prefix + url)
    const params = new URLSearchParams(queryparams)
    finalUrl.search = params.toString()
    console.log(finalUrl.toString())
    return finalUrl.toString()
}

const V1 = '/api/v1'

export async function getCategories(page, limit) {
    const offset = page * limit
    console.log(getBaseUrl(V1, '/categories', { limit, offset }))

    try {
        const response = await fetch(getBaseUrl(V1, '/categories', { limit, offset }))

        if (!response.ok) {
            return {
                ok: false,
                error: response.status
            }
        }

        const data = await response.json();
        console.log(data);

        return {
            ok: true,
            data: data
        };
    } catch (error) {
        console.error(error)
        return {
            ok: false,
            error: error
        };
    }
}


export async function createCategory(categoryName) {
    console.log(getBaseUrl(V1, '/categories'))

    try {
        const response = await fetch(getBaseUrl(V1, '/categories'), {
            method: 'POST',
            body: JSON.stringify({ name: categoryName })
        })

        if (!response.ok) {
            return {
                ok: false,
                error: response.status
            }
        }

        const data = await response.json();
        console.log(data);

        return {
            ok: true,
            data: data
        };
    } catch (error) {
        console.error(error)
        return {
            ok: false,
            error: error
        };
    }
}


export async function getProducts(limit, offset) {
    console.log(getBaseUrl(V1, '/products'))

    try {
        const response = await fetch(getBaseUrl(V1, '/products'))

        if (!response.ok) {
            return {
                ok: false,
                error: response.status
            }
        }

        const data = await response.json();
        console.log(data);

        return {
            ok: true,
            data: data
        };
    } catch (error) {
        console.error(error)
        return {
            ok: false,
            error: error
        };
    }
}

export async function deleteProduct(productId) {
    console.log(getBaseUrl(V1, '/products'), productId)

    try {
        const response = await fetch(getBaseUrl(V1, `/products/${productId}`), {
            method: 'DELETE'
        })

        if (!response.ok) {
            return {
                ok: false,
                error: response.status
            }
        }
        

        return {
            ok: true,
            data: 'Product deleted successfully'
        };
    } catch (error) {
        console.error(error)
        return {
            ok: false,
            error: error
        };
    }
}

export async function getProduct(productId) {
    console.log(getBaseUrl(V1, '/products'), productId)

    try {
        const response = await fetch(getBaseUrl(V1, `/products/${productId}`))

        if (!response.ok) {
            return {
                ok: false,
                error: response.status
            }
        }

        const data = await response.json();
        console.log(data);

        return {
            ok: true,
            data: data
        };
    } catch (error) {
        console.error(error)
        return {
            ok: false,
            error: error
        };
    }
}

export async function createProduct(productInfo) {

    try {
        const response = await fetch(getBaseUrl(V1, `/products/`), {
            method: 'POST',
            body: JSON.stringify(productInfo)
        })

        if (!response.ok) {
            return {
                ok: false,
                error: response.status
            }
        }

        const data = await response.json();
        console.log(data);

        return {
            ok: true,
            data: data
        };
    } catch (error) {
        console.error(error)
        return {
            ok: false,
            error: error
        };
    }
}

export async function updateProduct(productId, productInfo) {

    try {
        const response = await fetch(getBaseUrl(V1, `/products/${productId}`), {
            method: 'PUT',
            body: JSON.stringify(productInfo)
        })

        if (!response.ok) {
            return {
                ok: false,
                error: response.status
            }
        }

        const data = await response.json();
        console.log(data);

        return {
            ok: true,
            data: data
        };
    } catch (error) {
        console.error(error)
        return {
            ok: false,
            error: error
        };
    }
}

export async function linkCategory(productId, categoryId) {
    console.log(getBaseUrl(V1, `/products/${productId}/categories/link`), categoryId)

    try {
        const response = await fetch(getBaseUrl(V1, `/products/${productId}/categories/link`), {
            method: 'POST',
            body: JSON.stringify({ category_id: categoryId })
        })

        if (!response.ok) {
            return {
                ok: false,
                error: response.status
            }
        }

        return {
            ok: true,
            data: 'Category linked successfully'
        };
    } catch (error) {
        console.error(error)
        return {
            ok: false,
            error: error
        };
    }
}

export async function unlinkCategory(productId, categoryId) {
    console.log(getBaseUrl(V1, `/products/${productId}/categories/unlink`), categoryId)

    try {
        const response = await fetch(getBaseUrl(V1, `/products/${productId}/categories/unlink`), {
            method: 'POST',
            body: JSON.stringify({ category_id: categoryId })
        })

        if (!response.ok) {
            return {
                ok: false,
                error: response.status
            }
        }

        return {
            ok: true,
            data: 'Category unlinked successfully'
        };
    } catch (error) {
        console.error(error)
        return {
            ok: false,
            error: error
        };
    }
}


export async function searchCategory(categoryletter) {
    console.log(getBaseUrl(V1, '/categories/match', { name: categoryletter }))

    try {
        const response = await fetch(getBaseUrl(V1, `/categories/match`, { name: categoryletter }))

        if (!response.ok) {
            return {
                ok: false,
                error: response.status
            }
        }

        const data = await response.json();
        console.log(data, response.statusText, response.status);

        return {
            ok: true,
            data: data
        };
    } catch (error) {
        console.error(error)
        return {
            ok: false,
            error: error
        };
    }
}