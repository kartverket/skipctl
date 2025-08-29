local lib = import '../libsonnet/invalid.libsonnet';

{
  testing: 'hello',
  person: lib.makePerson('andreas', 25),
}
