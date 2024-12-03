import 'package:flutter/material.dart';
import 'package:high_seas_mobile/widgets/discover-movies.dart';

class DrawerWidget extends StatelessWidget {
  const DrawerWidget({super.key});

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: MediaQuery.of(context).size.width * 0.75, // 75% of screen will be occupied
      child: Drawer(
        backgroundColor: Color.fromRGBO(28, 0, 64, 100),
        // Your drawer content here
        child: ListView(
          padding: EdgeInsets.zero,
          children: [
            DrawerHeader(
              decoration: BoxDecoration(
                color: Theme.of(context).colorScheme.surface,
              ),
              child: Image.asset(
                "assets/high-seas-1.jpg",
                fit: BoxFit.fill,
              ),
            ),
            ListTile(
              tileColor: Theme.of(context).colorScheme.surface,
              title: Text(
                "Movies",
                style: TextStyle(
                  fontFamily: 'JetBrainsMono',
                  fontStyle: FontStyle.normal,
                  fontSize: 32,
                  color: Theme.of(context).colorScheme.onSurface,
                ),
              ),
            ),
            ListTile(
                tileColor: Theme.of(context).colorScheme.surface,
                title: Text(
                  "Discover Movies",
                  style: TextStyle(
                    fontFamily: 'JetBrainsMono',
                    fontStyle: FontStyle.normal,
                    fontSize: 18,
                    color: Theme.of(context).colorScheme.onSurface,
                  ),
                ),
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => const DiscoverMovies(title: 'High Seas'),
                    ),
                  );
                }
            ),
            ListTile(
                tileColor: Theme.of(context).colorScheme.surface,
                title: Text(
                  "Search Movies",
                  style: TextStyle(
                    fontFamily: 'JetBrainsMono',
                    fontStyle: FontStyle.normal,
                    fontSize: 18,
                    color: Theme.of(context).colorScheme.onSurface,
                  ),
                ),
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => Placeholder(),
                    ),
                  );
                }
            ),
            ListTile(
                tileColor: Theme.of(context).colorScheme.surface,
                title: Text(
                  "Now Playing Movies",
                  style: TextStyle(
                    fontFamily: 'JetBrainsMono',
                    fontStyle: FontStyle.normal,
                    fontSize: 18,
                    color: Theme.of(context).colorScheme.onSurface,
                  ),
                ),
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => Placeholder(),
                    ),
                  );
                }
            ),
            ListTile(
                tileColor: Theme.of(context).colorScheme.surface,
                title: Text(
                  "Popular Movies",
                  style: TextStyle(
                    fontFamily: 'JetBrainsMono',
                    fontStyle: FontStyle.normal,
                    fontSize: 18,
                    color: Theme.of(context).colorScheme.onSurface,
                  ),
                ),
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => Placeholder(),
                    ),
                  );
                }
            ),
            ListTile(
                tileColor: Theme.of(context).colorScheme.surface,
                title: Text(
                  "Top Rated Movies",
                  style: TextStyle(
                    fontFamily: 'JetBrainsMono',
                    fontStyle: FontStyle.normal,
                    fontSize: 18,
                    color: Theme.of(context).colorScheme.onSurface,
                  ),
                ),
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => Placeholder(),
                    ),
                  );
                }
            ),
            ListTile(
                tileColor: Theme.of(context).colorScheme.surface,
                title: Text(
                  "Upcoming Movies",
                  style: TextStyle(
                    fontFamily: 'JetBrainsMono',
                    fontStyle: FontStyle.normal,
                    fontSize: 18,
                    color: Theme.of(context).colorScheme.onSurface,
                  ),
                ),
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => Placeholder(),
                    ),
                  );
                }
            ),
            ListTile(
              tileColor: Theme.of(context).colorScheme.surface,
              title: Text(
                "Shows",
                style: TextStyle(
                  fontFamily: 'JetBrainsMono',
                  fontStyle: FontStyle.normal,
                  fontSize: 32,
                  color: Theme.of(context).colorScheme.onSurface,
                ),
              ),
            ),
            ListTile(
                tileColor: Theme.of(context).colorScheme.surface,
                title: Text(
                  "Discover Shows",
                  style: TextStyle(
                    fontFamily: 'JetBrainsMono',
                    fontStyle: FontStyle.normal,
                    fontSize: 18,
                    color: Theme.of(context).colorScheme.onSurface,
                  ),
                ),
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => Placeholder(),
                    ),
                  );
                }
            ),
            ListTile(
                tileColor: Theme.of(context).colorScheme.surface,
                title: Text(
                  "Search Shows",
                  style: TextStyle(
                    fontFamily: 'JetBrainsMono',
                    fontStyle: FontStyle.normal,
                    fontSize: 18,
                    color: Theme.of(context).colorScheme.onSurface,
                  ),
                ),
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => Placeholder(),
                    ),
                  );
                }
            ),
            ListTile(
                tileColor: Theme.of(context).colorScheme.surface,
                title: Text(
                  "Airing Today Shows",
                  style: TextStyle(
                    fontFamily: 'JetBrainsMono',
                    fontStyle: FontStyle.normal,
                    fontSize: 18,
                    color: Theme.of(context).colorScheme.onSurface,
                  ),
                ),
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => Placeholder(),
                    ),
                  );
                }
            ),
            ListTile(
                tileColor: Theme.of(context).colorScheme.surface,
                title: Text(
                  "Popular Shows",
                  style: TextStyle(
                    fontFamily: 'JetBrainsMono',
                    fontStyle: FontStyle.normal,
                    fontSize: 18,
                    color: Theme.of(context).colorScheme.onSurface,
                  ),
                ),
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => Placeholder(),
                    ),
                  );
                }
            ),
            ListTile(
                tileColor: Theme.of(context).colorScheme.surface,
                title: Text(
                  "Top Rated Shows",
                  style: TextStyle(
                    fontFamily: 'JetBrainsMono',
                    fontStyle: FontStyle.normal,
                    fontSize: 18,
                    color: Theme.of(context).colorScheme.onSurface,
                  ),
                ),
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => Placeholder(),
                    ),
                  );
                }
            ),
            ListTile(
                tileColor: Theme.of(context).colorScheme.surface,
                title: Text(
                  "On The Air Shows",
                  style: TextStyle(
                    fontFamily: 'JetBrainsMono',
                    fontStyle: FontStyle.normal,
                    fontSize: 18,
                    color: Theme.of(context).colorScheme.onSurface,
                  ),
                ),
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => Placeholder(),
                    ),
                  );
                }
            ),
          ],
        ),
      ),
    );
  }
}
