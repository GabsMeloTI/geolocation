import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    scenarios: {
        cep: {
            executor: 'constant-vus',
            vus: 25,
            duration: '30s',
            exec: 'testCep',
        },
        endereco: {
            executor: 'constant-vus',
            vus: 25,
            duration: '30s',
            exec: 'testEndereco',
        },
        state: {
            executor: 'constant-vus',
            vus: 25,
            duration: '30s',
            exec: 'testState',
        },
        calculateRoute: {
            executor: 'constant-vus',
            vus: 25,
            duration: '30s',
            exec: 'testCalculateRoute',
        },
        calculateCep: {
            executor: 'constant-vus',
            vus: 10,
            duration: '15s',
            exec: 'testCalculateCep',
        },
         calculateCoordinates: {
             executor: 'constant-vus',
             vus: 10,
             duration: '15s',
             exec: 'testCalculateCoordinates',
         },
        }
};

export function testCep() {
    const ceps = ["01001000", "20040002", "30140071"];
    const cep = ceps[Math.floor(Math.random() * ceps.length)];
    const url = `http://3.238.87.0:7070/address/find/${cep}`

    let res = http.get(url)

    check(res, {
        "cep status 200": (r) => r.status === 200,
    });

    sleep(1);
}

export function testEndereco() {
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
}

export function testState() {
    const stateRes = http.get('http://3.238.87.0:7070/address/state')
    check(stateRes, {
        'state status 200': (r) => r.status === 200,
    })

    sleep(1);
}

export function testCalculateRoute() {
    const calculateRoute = JSON.stringify({
        origin: "São Paulo, SP, Brasil",
        destination: "Rio de Janeiro, RJ, Brasil",
        consumptionCity: 8,
        consumptionHwy: 9.6,
        price: 5.7,
        axles: 4,
        type: "Truck",
        waypoints: [
            "Campinas, SP, Brasil"
        ],
        typeRoute: "RÁPIDA",
        route_options: {
            include_fuel_stations: true,
            include_route_map: true,
            include_toll_costs: true,
            include_weigh_stations: true,
            include_freight_calc: true,
            include_polyline: true
        }
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };

    const routeRes = http.post('http://3.238.87.0:7070/check-route-tolls-easy', calculateRoute, params)
    check(routeRes, {
        'calculate route status 200': (r) => r.status === 200,
    })

    sleep(1);
}

export function testCalculateCep() {
    const calculateCep = JSON.stringify({
            origin_cep: "01001-000",
            destination_cep: "90010-320",
            consumptionCity: 8,
            consumptionHwy: 9.6,
            price: 5.7,
            axles: 4,
            type: "Truck",
            waypoints: [
                "13015-000"
            ],
            typeRoute: "RÁPIDA",
            public_or_private: "private",
            favorite: false,
            route_options: {
                include_fuel_stations: true,
                include_route_map: true,
                include_toll_costs: true,
                include_weigh_stations: true,
                include_freight_calc: true,
                include_polyline: true
            },
            enterprise: false
        }
    );
    const params = {
        headers:{
            'Content-Type': 'application/json',
        }
    };

    const cepRes = http.post('http://3.238.87.0:7070/check-route-tolls-cep', calculateCep, params)
    check(cepRes, {
        'calculate cep status 200': (r) => r.status === 200,
    });

    sleep(1);
}

export function testCalculateCoordinates() {
    const calculateCoordinates = JSON.stringify({
        "origin_lat": "-23.451806",
        "origin_lng": "-46.712757",
        "destination_lat": "-30.035221",
        "destination_lng": "-51.226333",
        "consumptionCity": 8,
        "consumptionHwy": 9.6,
        "price": 5.7,
        "axles": 4,
        "type": "Truck",
        "waypoints": [
            {"lat": "-22.909756", "lng": "-47.05749"}
        ],
        "typeRoute": "RÁPIDA",
        "favorite": false,
        "route_options": {
            "include_fuel_stations": true,
            "include_route_map": true,
            "include_toll_costs": true,
            "include_weigh_stations": true,
            "include_freight_calc": true,
            "include_polyline": true
        }
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        }
    };

    const coordinatesRes = http.post('http://3.238.87.0:7070/check-route-tolls-coordinate', calculateCoordinates, params)
    check(coordinatesRes, {
        'calculate coordinates status 200': (r) => r.status === 200,
    });

    sleep(1);
}