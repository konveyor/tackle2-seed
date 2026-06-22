# tackle2-seed
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fkonveyor%2Ftackle2-seed.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fkonveyor%2Ftackle2-seed?ref=badge_shield)

Hub database seeds and seeding tools.

## Adding a New Target

1. Edit `resources/targets.yaml` and add a new label entry under an existing target, or add a new target item. Each label needs a `name` (display name) and `label` (must follow the `konveyor.io/target=<slug>` convention):

   ```yaml
   - name: OpenJDK 25
     label: konveyor.io/target=openjdk25
   ```

   Set `choice: true` on the target if it offers multiple version options.

2. Update the target's `description` to reflect the new option.

3. Bump the `version` field at the top of the file.

4. Run `make run-prepare` to auto-assign a UUID to any new items.

5. Commit and open a PR. Matching rulesets in [konveyor/rulesets](https://github.com/konveyor/rulesets) are synced separately by the `trigger-seed` workflow.

## Code of Conduct
Refer to Konveyor's Code of Conduct [here](https://github.com/konveyor/community/blob/main/CODE_OF_CONDUCT.md).


## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fkonveyor%2Ftackle2-seed.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fkonveyor%2Ftackle2-seed?ref=badge_large)