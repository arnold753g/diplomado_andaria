import '../core/attraction_format.dart';
import 'dart:async';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:url_launcher/url_launcher.dart';
import '../core/api.dart';
import '../core/models.dart';
import '../core/theme.dart';
import '../shared/widgets.dart';
import 'auth.dart';
import 'checkout.dart';

class ExplorePage extends StatefulWidget {
  const ExplorePage({super.key});
  @override
  State<ExplorePage> createState() => _ExplorePageState();
}

class _ExplorePageState extends State<ExplorePage> {
  final search = TextEditingController();
  Timer? debounce;
  bool attractions = false, busy = true;
  int page = 1, total = 0, generation = 0;
  String department = '', sort = 'next_departure';
  List<String> departments = [];
  List<Json> items = [];
  Object? error;
  @override
  void initState() {
    super.initState();
    load();
    options();
  }

  @override
  void dispose() {
    search.dispose();
    debounce?.cancel();
    super.dispose();
  }

  Future<void> options() async {
    try {
      final result = await context.read<AndariaApi>().request(
        '/packages/options',
      );
      if (mounted) {
        setState(() {
          departments = strings(result['departments']);
        });
      }
    } catch (_) {}
  }

  Future<void> load({bool more = false}) async {
    final current = ++generation;
    final next = more ? page + 1 : 1;
    setState(() {
      busy = true;
      error = null;
      if (!more) {
        items = [];
        total = 0;
      }
    });
    try {
      final result = await context.read<AndariaApi>().request(
        attractions ? '/attractions' : '/packages',
        query: {
          'page': next,
          'limit': 12,
          if (search.text.trim().isNotEmpty) 'search': search.text.trim(),
          if (department.isNotEmpty) 'department': department,
          if (!attractions) 'sort': sort,
        },
      );
      if (!mounted || current != generation) return;
      setState(() {
        items = [
          ...(more ? items : <Json>[]),
          ...objects(result[attractions ? 'attractions' : 'packages']),
        ];
        total = integer(result['pagination']['total']);
        page = next;
      });
    } catch (e) {
      if (mounted && current == generation) {
        setState(() {
          error = e;
        });
      }
    } finally {
      if (mounted && current == generation) {
        setState(() {
          busy = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) => SafeArea(
    child: RefreshIndicator(
      onRefresh: load,
      child: ListView(
        padding: const EdgeInsets.fromLTRB(20, 16, 20, 28),
        children: [
          Row(
            children: [
              Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: lime,
                  borderRadius: BorderRadius.circular(12),
                ),
                child: const Icon(Icons.explore_outlined),
              ),
              const SizedBox(width: 9),
              const Expanded(
                child: Text(
                  'andaria',
                  style: TextStyle(
                    fontSize: 25,
                    fontWeight: FontWeight.w700,
                    letterSpacing: -1,
                  ),
                ),
              ),
              const SizedBox(width: 12),
              const Flexible(
                child: Text(
                  'BOLIVIA',
                  style: TextStyle(
                    fontSize: 11,
                    letterSpacing: 2,
                    fontWeight: FontWeight.w600,
                    color: muted,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 30),
          Text(
            'Sal de la rutina.\nDescubre Bolivia.',
            style: Theme.of(context).textTheme.headlineLarge,
          ),
          const SizedBox(height: 12),
          const Text(
            'Tu próxima experiencia, comenzando por Tarija.',
            style: TextStyle(color: muted),
          ),
          const SizedBox(height: 24),
          TextField(
            controller: search,
            decoration: InputDecoration(
              hintText: '¿Qué te gustaría explorar?',
              prefixIcon: const Icon(Icons.search),
              suffixIcon: search.text.isEmpty
                  ? null
                  : IconButton(
                      tooltip: 'Limpiar búsqueda',
                      icon: const Icon(Icons.close),
                      onPressed: () {
                        search.clear();
                        debounce?.cancel();
                        load();
                      },
                    ),
            ),
            onChanged: (_) {
              setState(() {});
              debounce?.cancel();
              debounce = Timer(const Duration(milliseconds: 400), load);
            },
          ),
          const SizedBox(height: 16),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: [
              ChoiceChip(
                label: const Text('Experiencias'),
                selected: !attractions,
                onSelected: (_) {
                  setState(() {
                    attractions = false;
                  });
                  load();
                },
              ),
              ChoiceChip(
                label: const Text('Lugares'),
                selected: attractions,
                onSelected: (_) {
                  setState(() {
                    attractions = true;
                  });
                  load();
                },
              ),
            ],
          ),
          const SizedBox(height: 14),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: [
              PopupMenuButton<String>(
                initialValue: department,
                onSelected: (v) {
                  setState(() {
                    department = v;
                  });
                  load();
                },
                itemBuilder: (_) => [
                  const PopupMenuItem(value: '', child: Text('Toda Bolivia')),
                  ...departments.map(
                    (v) => PopupMenuItem(value: v, child: Text(v)),
                  ),
                ],
                child: Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 14,
                    vertical: 12,
                  ),
                  decoration: BoxDecoration(
                    border: Border.all(color: line),
                    borderRadius: BorderRadius.circular(24),
                  ),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const Icon(Icons.location_on_outlined, size: 18),
                      const SizedBox(width: 6),
                      Text(department.isEmpty ? 'Toda Bolivia' : department),
                      const Icon(Icons.expand_more, size: 18),
                    ],
                  ),
                ),
              ),
              if (!attractions)
                PopupMenuButton<String>(
                  initialValue: sort,
                  tooltip: 'Ordenar experiencias',
                  onSelected: (v) {
                    setState(() {
                      sort = v;
                    });
                    load();
                  },
                  itemBuilder: (_) => const [
                    PopupMenuItem(
                      value: 'next_departure',
                      child: Text('Próximas salidas'),
                    ),
                    PopupMenuItem(
                      value: 'price_asc',
                      child: Text('Menor precio'),
                    ),
                    PopupMenuItem(
                      value: 'price_desc',
                      child: Text('Mayor precio'),
                    ),
                    PopupMenuItem(
                      value: 'newest',
                      child: Text('Más recientes'),
                    ),
                  ],
                  child: Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 14,
                      vertical: 12,
                    ),
                    decoration: BoxDecoration(
                      border: Border.all(color: line),
                      borderRadius: BorderRadius.circular(24),
                    ),
                    child: const Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(Icons.tune, size: 18),
                        SizedBox(width: 6),
                        Text('Ordenar'),
                      ],
                    ),
                  ),
                ),
            ],
          ),
          SectionTitle(
            attractions ? 'Lugares que inspiran' : 'Experiencias para recordar',
          ),
          if (!busy && error == null)
            Padding(
              padding: const EdgeInsets.only(bottom: 16),
              child: Text(
                '$total ${attractions ? 'lugares' : 'experiencias'} disponibles',
                style: const TextStyle(color: muted),
              ),
            ),
          ...items.map(
            (item) => attractions
                ? AttractionCard(Attraction(item))
                : PackageCard(TravelPackage(item)),
          ),
          if (busy)
            const Padding(
              padding: EdgeInsets.all(32),
              child: Center(child: CircularProgressIndicator()),
            ),
          if (error != null)
            EmptyView(
              icon: Icons.wifi_off_outlined,
              title: 'No pudimos cargar el catálogo',
              description: error is ApiFailure
                  ? (error as ApiFailure).message
                  : 'Inténtalo nuevamente.',
              action: 'Volver a intentar',
              onAction: () => load(more: items.isNotEmpty),
            ),
          if (!busy && error == null && items.isEmpty)
            const EmptyView(
              icon: Icons.travel_explore,
              title: 'Todavía no encontramos resultados',
              description: 'Prueba otra búsqueda o cambia el departamento.',
            ),
          if (!busy && error == null && items.length < total)
            OutlinedButton(
              onPressed: () => load(more: true),
              child: const Text('Ver más resultados'),
            ),
        ],
      ),
    ),
  );
}

class PackageCard extends StatelessWidget {
  const PackageCard(this.item, {super.key});
  final TravelPackage item;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(bottom: 20),
    child: Material(
      color: Colors.white,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(18),
        side: const BorderSide(color: line),
      ),
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: () => openPage(context, PackageDetailPage(id: item.id)),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            TravelPhoto(
              type: 'packages',
              id: item.id,
              photos: item.photos.take(1).toList(),
            ),
            Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    item.location,
                    style: const TextStyle(color: muted, fontSize: 12),
                  ),
                  const SizedBox(height: 6),
                  Text(
                    item.name,
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  const SizedBox(height: 6),
                  Text(
                    '${item.duration} · ${item.s('agency_name')}',
                    style: const TextStyle(color: muted),
                  ),
                  const SizedBox(height: 14),
                  Wrap(
                    spacing: 12,
                    runSpacing: 8,
                    crossAxisAlignment: WrapCrossAlignment.center,
                    children: [
                      Text(
                        'Desde ${money(item.price)}',
                        style: const TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                      Text(
                        item.bookable
                            ? 'Salidas disponibles'
                            : 'Sin venta abierta',
                        style: TextStyle(
                          fontSize: 12,
                          color: item.bookable
                              ? const Color(0xFF466000)
                              : muted,
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    ),
  );
}

class AttractionCard extends StatelessWidget {
  const AttractionCard(this.item, {super.key});
  final Attraction item;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(bottom: 20),
    child: Material(
      color: Colors.white,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(18),
        side: const BorderSide(color: line),
      ),
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: () => openPage(context, AttractionDetailPage(id: item.id)),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            TravelPhoto(
              type: 'attractions',
              id: item.id,
              photos: item.photos.take(1).toList(),
            ),
            Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    item.location,
                    style: const TextStyle(color: muted, fontSize: 12),
                  ),
                  const SizedBox(height: 6),
                  Text(
                    item.name,
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  const SizedBox(height: 8),
                  Text(
                    item.n('admission_cents') == 0
                        ? 'Ingreso gratuito'
                        : 'Ingreso: ${money(item.n('admission_cents'))}',
                    style: const TextStyle(color: muted),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    ),
  );
}

class PackageDetailPage extends StatelessWidget {
  const PackageDetailPage({super.key, required this.id});
  final int id;
  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: const Text('La experiencia')),
    body: AsyncContent<TravelPackage>(
      load: () async => TravelPackage(
        Map<String, dynamic>.from(
          await context.read<AndariaApi>().request('/packages/$id') as Map,
        ),
      ),
      builder: (item, reload) => RefreshIndicator(
        onRefresh: reload,
        child: ListView(
          children: [
            TravelPhoto(
              type: 'packages',
              id: id,
              photos: item.photos,
              height: 280,
              allowZoom: true,
            ),
            Padding(
              padding: const EdgeInsets.all(20),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(item.location, style: const TextStyle(color: muted)),
                  const SizedBox(height: 8),
                  Text(
                    item.name,
                    style: Theme.of(context).textTheme.headlineMedium,
                  ),
                  Fact(Icons.schedule, item.duration),
                  Fact(Icons.storefront_outlined, item.s('agency_name')),
                  if (item.s('difficulty').isNotEmpty)
                    Fact(
                      Icons.hiking,
                      'Dificultad: ${{'easy': 'fácil', 'moderate': 'moderada', 'demanding': 'exigente'}[item.s('difficulty')]}',
                    ),
                  const SizedBox(height: 12),
                  Text(
                    'Desde ${money(item.price)} por viajero que paga',
                    style: Theme.of(context).textTheme.titleLarge,
                  ),
                  if (item.n('foreign_surcharge_cents') > 0)
                    Text(
                      'Adicional por extranjero: ${money(item.n('foreign_surcharge_cents'))}',
                      style: const TextStyle(color: muted),
                    ),
                  const SectionTitle('Acerca de esta experiencia'),
                  Text(item.s('description')),
                  ...['includes', 'excludes', 'bring']
                      .where((key) => strings(item.json[key]).isNotEmpty)
                      .expand(
                        (key) => [
                          SectionTitle(
                            {
                              'includes': 'Qué incluye',
                              'excludes': 'Qué no incluye',
                              'bring': 'Qué llevar',
                            }[key]!,
                          ),
                          ...strings(item.json[key]).map(
                            (text) => Fact(
                              key == 'includes'
                                  ? Icons.check
                                  : Icons.circle_outlined,
                              text,
                            ),
                          ),
                        ],
                      ),
                  const SectionTitle('Tu itinerario'),
                  ...item
                      .list('itinerary')
                      .map(
                        (day) => ExpansionTile(
                          tilePadding: EdgeInsets.zero,
                          title: Text(
                            'Día ${day['day_number']} · ${day['title']}',
                          ),
                          children: [
                            Align(
                              alignment: Alignment.centerLeft,
                              child: Padding(
                                padding: const EdgeInsets.only(bottom: 16),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(day['description']?.toString() ?? ''),
                                    ...strings(day['activities']).map(
                                      (text) => Fact(Icons.arrow_forward, text),
                                    ),
                                    ...objects(day['attractions']).map(
                                      (a) => ListTile(
                                        contentPadding: EdgeInsets.zero,
                                        leading: const Icon(
                                          Icons.place_outlined,
                                        ),
                                        title: Text(
                                          a['attraction']?['name']
                                                  ?.toString() ??
                                              'Atracción del itinerario',
                                        ),
                                        trailing:
                                            integer(a['attraction']?['id']) > 0
                                            ? const Icon(Icons.chevron_right)
                                            : null,
                                        onTap:
                                            integer(a['attraction']?['id']) > 0
                                            ? () => openPage(
                                                context,
                                                AttractionDetailPage(
                                                  id: integer(
                                                    a['attraction']['id'],
                                                  ),
                                                ),
                                              )
                                            : null,
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                  const SectionTitle('Cancelaciones'),
                  Text(
                    item.b('cancellation_allowed')
                        ? 'Puedes cancelar con al menos ${item.n('cancellation_notice_hours')} horas de anticipación a la salida, según la política del paquete.'
                        : 'Este paquete no permite cancelaciones del turista.',
                  ),
                  const SectionTitle('Elige una salida'),
                  const Text(
                    'Todos los horarios corresponden a Bolivia (UTC−4).',
                    style: TextStyle(color: muted),
                  ),
                  const SizedBox(height: 12),
                  if (item.departures.isEmpty)
                    const Notice(
                      'Todavía no hay salidas publicadas para esta experiencia.',
                    ),
                  ...item.departures.map(
                    (departure) => Container(
                      margin: const EdgeInsets.only(bottom: 12),
                      padding: const EdgeInsets.all(16),
                      decoration: BoxDecoration(
                        border: Border.all(color: line),
                        borderRadius: BorderRadius.circular(16),
                      ),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            dateLabel(departure.s('starts_at')),
                            style: Theme.of(context).textTheme.titleMedium,
                          ),
                          Fact(
                            Icons.people_outline,
                            countLabel(
                              departure.n('available_capacity'),
                              'cupo disponible',
                              'cupos disponibles',
                            ),
                          ),
                          Fact(
                            Icons.location_on_outlined,
                            departure.s('meeting_point'),
                          ),
                          if (!departure.b('minimum_reached'))
                            Notice(
                              'La salida necesita ${departure.n('remaining_for_minimum')} cupos confirmados más para alcanzar el mínimo.',
                              warning: true,
                            ),
                          const SizedBox(height: 8),
                          SizedBox(
                            width: double.infinity,
                            child: FilledButton(
                              onPressed: departure.b('bookable')
                                  ? () async {
                                      if (!await requireTourist(context) ||
                                          !context.mounted ||
                                          ModalRoute.of(context)?.isCurrent ==
                                              false) {
                                        return;
                                      }
                                      await openPage(
                                        context,
                                        CheckoutPage(
                                          item: item,
                                          departure: departure,
                                        ),
                                      );
                                      await reload();
                                    }
                                  : null,
                              child: Text(
                                departure.b('bookable')
                                    ? 'Elegir esta salida'
                                    : 'Compra no disponible',
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    ),
  );
}

class AttractionDetailPage extends StatefulWidget {
  const AttractionDetailPage({super.key, required this.id});
  final int id;
  @override
  State<AttractionDetailPage> createState() => _AttractionDetailPageState();
}

class _AttractionDetailPageState extends State<AttractionDetailPage> {
  bool favorite = false, working = false;
  Object? favoriteAccount;
  int favoriteGeneration = 0;
  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final current = context.watch<AndariaApi>().user?['id'];
    if (current != favoriteAccount) {
      favoriteAccount = current;
      favorite = false;
      checkFavorite();
    }
  }

  Future<void> checkFavorite() async {
    final generation = ++favoriteGeneration;
    final api = context.read<AndariaApi>();
    if (!api.signedIn) return;
    try {
      final result = await api.request(
        '/me/favorites/${widget.id}',
        authenticated: true,
      );
      if (mounted && generation == favoriteGeneration) {
        setState(() {
          favorite = result['favorite'] == true;
        });
      }
    } catch (_) {}
  }

  Future<void> toggle() async {
    if (!await requireTourist(context) || !mounted || working) return;
    favoriteGeneration++;
    setState(() {
      working = true;
    });
    try {
      final data = await context.read<AndariaApi>().request(
        '/me/favorites/${widget.id}',
        method: favorite ? 'DELETE' : 'PUT',
        authenticated: true,
      );
      if (mounted) {
        setState(() {
          favorite = data['favorite'] == true;
        });
      }
    } catch (e) {
      if (mounted) message(context, e);
    } finally {
      if (mounted) {
        setState(() {
          working = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(
      title: const Text('Descubre el lugar'),
      actions: [
        IconButton(
          tooltip: favorite ? 'Quitar de favoritos' : 'Guardar en favoritos',
          onPressed: working ? null : toggle,
          icon: Icon(favorite ? Icons.favorite : Icons.favorite_border),
        ),
      ],
    ),
    body: AsyncContent<Attraction>(
      load: () async => Attraction(
        Map<String, dynamic>.from(
          await context.read<AndariaApi>().request('/attractions/${widget.id}')
              as Map,
        ),
      ),
      builder: (item, reload) => RefreshIndicator(
        onRefresh: reload,
        child: ListView(
          children: [
            TravelPhoto(
              type: 'attractions',
              id: item.id,
              photos: item.photos,
              height: 280,
              allowZoom: true,
            ),
            Padding(
              padding: const EdgeInsets.all(20),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(item.location, style: const TextStyle(color: muted)),
                  const SizedBox(height: 8),
                  Text(
                    item.name,
                    style: Theme.of(context).textTheme.headlineMedium,
                  ),
                  Fact(Icons.location_on_outlined, item.s('address')),
                  Fact(
                    Icons.payments_outlined,
                    item.n('admission_cents') == 0
                        ? 'Ingreso gratuito'
                        : money(item.n('admission_cents')),
                  ),
                  const SectionTitle('Acerca del lugar'),
                  Text(item.s('description')),
                  const SectionTitle('Horarios y temporada'),
                  Text(attractionHours(item)),
                  Text(attractionSeason(item)),
                  if (item.s('recommendations').isNotEmpty) ...[
                    const SectionTitle('Antes de ir'),
                    Text(item.s('recommendations')),
                  ],
                  if (item.json['latitude'] != null &&
                      item.json['longitude'] != null) ...[
                    const SizedBox(height: 24),
                    OutlinedButton.icon(
                      onPressed: () async {
                        final uri = Uri.https('www.google.com', '/maps/search/', {
                          'api': '1',
                          'query':
                              '${item.json['latitude']},${item.json['longitude']}',
                        });
                        try {
                          if (!await launchUrl(
                                uri,
                                mode: LaunchMode.externalApplication,
                              ) &&
                              context.mounted) {
                            message(
                              context,
                              const ApiFailure('No se pudo abrir el mapa.'),
                            );
                          }
                        } catch (e) {
                          if (context.mounted) message(context, e);
                        }
                      },
                      icon: const Icon(Icons.map_outlined),
                      label: const Text('Abrir ubicación en el mapa'),
                    ),
                  ],
                ],
              ),
            ),
          ],
        ),
      ),
    ),
  );
}

class FavoritesPage extends StatefulWidget {
  const FavoritesPage({super.key});
  @override
  State<FavoritesPage> createState() => _FavoritesPageState();
}

class _FavoritesPageState extends State<FavoritesPage> {
  int page = 1;
  @override
  Widget build(BuildContext context) {
    final api = context.watch<AndariaApi>();
    return Scaffold(
      appBar: AppBar(title: const Text('Tus lugares favoritos')),
      body: !api.signedIn
          ? EmptyView(
              icon: Icons.favorite_border,
              title: 'Guarda lo que te inspira',
              description:
                  'Inicia sesión para guardar atracciones y encontrarlas cuando quieras.',
              action: 'Iniciar sesión',
              onAction: () => requireTourist(context),
            )
          : AsyncContent<Json>(
              key: ValueKey('$page-${api.user?['id']}'),
              load: () async => Map<String, dynamic>.from(
                await api.request(
                      '/me/favorites',
                      query: {'page': page, 'limit': 20},
                      authenticated: true,
                    )
                    as Map,
              ),
              builder: (data, reload) {
                final items = objects(
                  data['attractions'],
                ).map(Attraction.new).toList();
                if (items.isEmpty) {
                  return EmptyView(
                    icon: Icons.favorite_border,
                    title: 'Tu lista empieza con un lugar',
                    description:
                        'Explora una atracción y toca el corazón para guardarla.',
                    action: page > 1 ? 'Volver a la primera página' : null,
                    onAction: () => setState(() {
                      page = 1;
                    }),
                  );
                }
                return RefreshIndicator(
                  onRefresh: reload,
                  child: ListView(
                    padding: const EdgeInsets.all(20),
                    children: [
                      ...items.map(
                        (item) => Padding(
                          padding: const EdgeInsets.only(bottom: 8),
                          child: Column(
                            children: [
                              AttractionCard(item),
                              Align(
                                alignment: Alignment.centerRight,
                                child: TextButton.icon(
                                  onPressed: () async {
                                    try {
                                      await api.request(
                                        '/me/favorites/${item.id}',
                                        method: 'DELETE',
                                        authenticated: true,
                                      );
                                      await reload();
                                    } catch (e) {
                                      if (context.mounted) message(context, e);
                                    }
                                  },
                                  icon: const Icon(Icons.favorite_border),
                                  label: const Text('Quitar de favoritos'),
                                ),
                              ),
                            ],
                          ),
                        ),
                      ),
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          TextButton(
                            onPressed: page > 1
                                ? () => setState(() {
                                    page--;
                                  })
                                : null,
                            child: const Text('Anterior'),
                          ),
                          Text('Página $page'),
                          TextButton(
                            onPressed:
                                page * 20 < integer(data['pagination']['total'])
                                ? () => setState(() {
                                    page++;
                                  })
                                : null,
                            child: const Text('Siguiente'),
                          ),
                        ],
                      ),
                    ],
                  ),
                );
              },
            ),
    );
  }
}
