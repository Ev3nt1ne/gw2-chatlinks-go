"""Regenerates realworld_fixtures.json from the live GW2 API.

Run from this directory: python3 gather_fixtures.py

Pulls real, authoritative (id, chat_link) pairs directly from /v2/items,
/v2/skills, and /v2/recipes (all three expose a `chat_link` field — verified
against the live API, not assumed). /v2/achievements and /v2/continents
(map points of interest) don't expose chat_link, so those fixtures carry
only a real `id` with no `code` field; realworld_test.go treats those as
self-consistency checks (encode then decode recovers the same ID) rather
than external ground-truth matches.

This is intentionally not run automatically by `go test` or CI — it's a
one-off/occasional refresh tool. Re-run it (and review the diff) if you
want to grow or refresh the sample set, e.g. when adding coverage for a
profession/item-type combination not yet represented.
"""

import json
import random
import time
import urllib.request

BASE = "https://api.guildwars2.com/v2"


def get(path):
    with urllib.request.urlopen(f"{BASE}{path}", timeout=30) as resp:
        return json.load(resp)


def bulk_fetch(endpoint, ids, batch_size=200):
    results = []
    for i in range(0, len(ids), batch_size):
        batch = ids[i : i + batch_size]
        idstr = ",".join(str(x) for x in batch)
        results.extend(get(f"/{endpoint}?ids={idstr}&v=latest"))
        time.sleep(0.2)
    return results


def main():
    random.seed(42)
    fixtures = []

    item_ids = random.sample(get("/items"), 90)
    skill_ids = random.sample(get("/skills"), 60)
    recipe_ids = random.sample(get("/recipes"), 60)

    for it in bulk_fetch("items", item_ids):
        if "chat_link" in it:
            fixtures.append({"type": "item", "code": it["chat_link"], "id": it["id"], "name": it.get("name", "")})
    for sk in bulk_fetch("skills", skill_ids):
        if "chat_link" in sk:
            fixtures.append({"type": "skill", "code": sk["chat_link"], "id": sk["id"], "name": sk.get("name", "")})
    for r in bulk_fetch("recipes", recipe_ids):
        if "chat_link" in r:
            fixtures.append({"type": "recipe", "code": r["chat_link"], "id": r["id"], "name": ""})

    random.seed(7)
    ach_ids = random.sample(get("/achievements"), 15)
    for a in get("/achievements?ids=" + ",".join(str(i) for i in ach_ids) + "&v=latest"):
        fixtures.append({"type": "achievement", "id": a["id"], "name": a.get("name", "")})

    floor = get("/continents/1/floors/1?v=latest")
    pois = []
    for region in floor["regions"].values():
        for mp in region["maps"].values():
            pois.extend(mp.get("points_of_interest", {}).values())
    for p in random.sample(pois, min(15, len(pois))):
        fixtures.append({"type": "map", "id": p["id"], "name": p.get("name", p.get("type", ""))})

    print(f"collected {len(fixtures)} fixtures")
    with open("realworld_fixtures.json", "w") as f:
        json.dump(fixtures, f, indent=2)


if __name__ == "__main__":
    main()
