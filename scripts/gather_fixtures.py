"""Regenerates chatlinks/testdata/realworld_fixtures.json from the live GW2 API.

Run from anywhere: python3 scripts/gather_fixtures.py

Samples by *category*, not raw random count — the goal is coverage across
known subtypes (item types including dyes, map point types, skills spread
across professions, recipe types), not maximizing total entries. See
docs/GW2_CHATLINKS_GO_REVIEW_PLAN.md (HeroAscent workspace, not this repo)
for the verified category facts this script's grouping logic is based on.

Pulls real, authoritative (id, chat_link) pairs directly from /v2/items,
/v2/skills, and /v2/recipes (all three expose a `chat_link` field —
verified against the live API, not assumed). /v2/achievements and
/v2/continents (map points of interest) don't expose chat_link, so those
fixtures carry only a real `id` with no `code` field; realworld_test.go
treats those as self-consistency checks (encode then decode recovers the
same ID) rather than external ground-truth matches.

This is intentionally not run automatically by `go test` or CI — it's a
one-off/occasional refresh tool. Re-run it (and review the diff) if you
want to grow or rebalance the sample set.
"""

import json
import os
import random
import time
import urllib.request

BASE = "https://api.guildwars2.com/v2"
OUT_PATH = os.path.join(
    os.path.dirname(__file__), "..", "chatlinks", "testdata", "realworld_fixtures.json"
)

PER_GROUP = 8


def get(path):
    with urllib.request.urlopen(f"{BASE}{path}", timeout=30) as resp:
        return json.load(resp)


def bulk_fetch(endpoint, ids, batch_size=200, delay=0.15):
    results = []
    for i in range(0, len(ids), batch_size):
        batch = ids[i : i + batch_size]
        idstr = ",".join(str(x) for x in batch)
        results.extend(get(f"/{endpoint}?ids={idstr}&v=latest"))
        time.sleep(delay)
    return results


def sample_per_group(groups, rng, n=PER_GROUP):
    chosen = []
    for key in sorted(groups.keys(), key=str):
        bucket = groups[key]
        chosen.extend((key, x) for x in rng.sample(bucket, min(n, len(bucket))))
    return chosen


def gather_items(rng):
    all_ids = get("/items")
    sample_ids = rng.sample(all_ids, min(15000, len(all_ids)))
    items = [it for it in bulk_fetch("items", sample_ids) if "chat_link" in it]

    groups = {}
    for it in items:
        details = it.get("details") or {}
        if it.get("type") == "Consumable" and details.get("unlock_type") == "Dye":
            key = "Dye"
        else:
            key = it.get("type", "Unknown")
        groups.setdefault(key, []).append(it)

    fixtures = []
    for category, it in sample_per_group(groups, rng):
        fixtures.append(
            {
                "type": "item",
                "category": category,
                "code": it["chat_link"],
                "id": it["id"],
                "name": it.get("name", ""),
            }
        )
    print(f"items: {len(groups)} categories found ({sorted(groups.keys())}), {len(fixtures)} sampled")
    return fixtures


def gather_skills(rng):
    all_ids = get("/skills")
    sample_ids = rng.sample(all_ids, min(8000, len(all_ids)))
    skills = [sk for sk in bulk_fetch("skills", sample_ids) if "chat_link" in sk]

    groups = {}
    for sk in skills:
        professions = sk.get("professions") or []
        key = professions[0] if professions else "(none)"
        groups.setdefault(key, []).append(sk)

    fixtures = []
    for category, sk in sample_per_group(groups, rng):
        fixtures.append(
            {
                "type": "skill",
                "category": category,
                "code": sk["chat_link"],
                "id": sk["id"],
                "name": sk.get("name", ""),
            }
        )
    print(f"skills: {len(groups)} profession groups found ({sorted(groups.keys())}), {len(fixtures)} sampled")
    return fixtures


def gather_recipes(rng):
    all_ids = get("/recipes")
    sample_ids = rng.sample(all_ids, min(4000, len(all_ids)))
    recipes = [r for r in bulk_fetch("recipes", sample_ids) if "chat_link" in r]

    groups = {}
    for r in recipes:
        groups.setdefault(r.get("type", "Unknown"), []).append(r)

    # Recipe `type` is a crafting-category label only — it doesn't change
    # how the chat link decodes (unlike item type, where dyes are a
    # genuinely distinct case). With 50+ distinct types, sampling at the
    # same per-group count as items would bloat the fixture set for very
    # little marginal verification value, so use a smaller count here.
    fixtures = []
    for category, r in sample_per_group(groups, rng, n=3):
        fixtures.append(
            {
                "type": "recipe",
                "category": category,
                "code": r["chat_link"],
                "id": r["id"],
                "name": "",
            }
        )
    print(f"recipes: {len(groups)} type groups found ({sorted(groups.keys())}), {len(fixtures)} sampled")
    return fixtures


def gather_achievements(rng):
    all_ids = get("/achievements")
    sample_ids = rng.sample(all_ids, min(500, len(all_ids)))
    achievements = bulk_fetch("achievements", sample_ids)

    groups = {}
    for a in achievements:
        flags = a.get("flags") or []
        key = "Repeatable" if "Repeatable" in flags else "Permanent" if "Permanent" in flags else "(other)"
        groups.setdefault(key, []).append(a)

    fixtures = []
    for category, a in sample_per_group(groups, rng, n=8):
        fixtures.append(
            {"type": "achievement", "category": category, "id": a["id"], "name": a.get("name", "")}
        )
    print(f"achievements: {len(groups)} groups found ({sorted(groups.keys())}), {len(fixtures)} sampled")
    return fixtures


def gather_map_points(rng):
    # Spread across multiple floors of both continents (1 = Tyria, 2 =
    # Mists), not just floor 1, so newer/lobby-adjacent maps are
    # represented too, not just core Tyria.
    floor_specs = [(1, 1), (1, 30), (1, 50), (1, 71), (2, 1)]
    groups = {}
    for continent_id, floor_id in floor_specs:
        try:
            floor = get(f"/continents/{continent_id}/floors/{floor_id}?v=latest")
        except Exception as e:
            print(f"  skipping continent {continent_id} floor {floor_id}: {e}")
            continue
        for region in floor.get("regions", {}).values():
            for mp in region.get("maps", {}).values():
                for poi in mp.get("points_of_interest", {}).values():
                    key = poi.get("type", "Unknown")
                    groups.setdefault(key, []).append(poi)

    fixtures = []
    for category, poi in sample_per_group(groups, rng):
        fixtures.append(
            {"type": "map", "category": category, "id": poi["id"], "name": poi.get("name", category)}
        )
    print(f"map points: {len(groups)} type groups found ({sorted(groups.keys())}), {len(fixtures)} sampled")
    return fixtures


def main():
    rng = random.Random(42)
    fixtures = []
    fixtures += gather_items(rng)
    fixtures += gather_skills(rng)
    fixtures += gather_recipes(rng)
    fixtures += gather_achievements(rng)
    fixtures += gather_map_points(rng)

    print(f"\ntotal: {len(fixtures)} fixtures")
    with open(OUT_PATH, "w") as f:
        json.dump(fixtures, f, indent=2)
    print(f"wrote {OUT_PATH}")


if __name__ == "__main__":
    main()
