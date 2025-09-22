import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    vus: 100,
    duration: '30s',
};

export default function() {
    //cep
    const ceps = ["01001000", "20040002", "30140071"];
    const cep = ceps[Math.floor(Math.random() * ceps.length)];
    const url = `http://3.238.87.0:7070/address/find/${cep}`;

    let res = http.get(url)

    check(res, {
        "cep status 200": (r) => r.status === 200,
    });

    sleep(1);

    // endereco completo (rua, bairro, numero)

    const enderecos = [
        { rua: "Avenida Paulista", bairro: "Bela Vista", numero: "1000" },
        { rua: "Rua das Flores", bairro: "Jardim América", numero: "500" },
        { rua: "Avenida Brasil", bairro: "Centro", numero: "200" }
    ];
    const endereco = enderecos[Math.floor(Math.random() * enderecos.length)];
    const query = encodeURIComponent(`${endereco.rua}, ${endereco.bairro}, ${endereco.numero}`);

    const enderecoRes = http.get(
        `http://3.238.87.0:7070/address/find/v2?q=${query}`
    )
    check(enderecoRes, {
        "endereco status 200": (r) => r.status === 200,
    });

    sleep(1);

    // estados
    const stateRes = http.get('http://3.238.87.0:7070/address/state')
    check(stateRes, {
        'state status 200': (r) => r.status === 200,
    })

    sleep(2);
}